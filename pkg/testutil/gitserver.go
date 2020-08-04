// The code for the gitserver is assembled from: github.com/dcu/git-http-server
// So their license applies.

package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubism/smorgasbord/pkg/util"
)

type gitRouteFunc func(route *gitRoute, w http.ResponseWriter, r *http.Request)

type gitRouteMatcher struct {
	Matcher *regexp.Regexp
	Params  []string
	Handler gitRouteFunc
}

type gitRoute struct {
	RepoPath     string
	File         string
	MatchedRoute gitRouteMatcher
}

func (route *gitRoute) Dispatch(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	route.MatchedRoute.Handler(route, w, r)
}

type gitCommand struct {
	ProcInput *bytes.Reader
	Args      []string
}

func (c *gitCommand) Run(wait bool) (io.ReadCloser, error) {
	cmd := exec.Command("git", c.Args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if c.ProcInput != nil {
		cmd.Stdin = c.ProcInput
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if wait {
		err = cmd.Wait()
		if err != nil {
			return nil, err
		}
	}
	return stdout, nil
}

func (c *gitCommand) RunAndGetOutput() []byte {
	stdout, err := c.Run(false)
	if err != nil {
		return []byte{}
	}
	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		return []byte{}
	}
	return data
}

func writeGitToHTTP(w http.ResponseWriter, c gitCommand) {
	stdout, err := c.Run(false)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	_, err = io.Copy(w, stdout)
	if err != nil {
		panic(err)
	}
}

type GitServer struct {
	server    *http.Server
	serverLis net.Listener
	rootDir   string
	routes    []gitRouteMatcher
}

func NewGitServer(rootDir string) (*GitServer, error) {
	port, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	g := &GitServer{rootDir: rootDir}
	g.routes = []gitRouteMatcher{
		{Matcher: regexp.MustCompile("(.*?)/info/refs$"), Handler: g.getInfoRefs},
		{Matcher: regexp.MustCompile("(.*?)/git-upload-pack$"), Handler: g.uploadPack},
		{Matcher: regexp.MustCompile("(.*?)/git-receive-pack$"), Handler: g.receivePack},
		{Matcher: regexp.MustCompile("(.*)"), Params: []string{"go-get"}, Handler: g.goGettable},
	}
	g.server = &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parsedRoute := g.matchRoute(r)
		if parsedRoute != nil {
			parsedRoute.Dispatch(w, r)
		} else {
			fmt.Fprintf(w, "nothing to see here\n")
		}
	})}
	g.serverLis, err = net.Listen("tcp", g.server.Addr)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := g.server.Serve(g.serverLis); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return g, nil
}

func (g *GitServer) Close() error {
	_ = g.server.Close()
	return g.serverLis.Close()
}

func (g *GitServer) GetAddr() string {
	return g.server.Addr
}

func (g *GitServer) matchRoute(r *http.Request) *gitRoute {
	path := r.URL.Path[1:]
	for _, routeMatcher := range g.routes {
		matches := routeMatcher.Matcher.FindStringSubmatch(path)
		if matches != nil && g.areParamsMatched(r.URL.Query(), &routeMatcher) {
			repoName := matches[1]
			file := strings.Replace(path, repoName+"/", "", 1)
			return &gitRoute{RepoPath: repoName, File: file, MatchedRoute: routeMatcher}
		}
	}
	return nil
}

func (g *GitServer) areParamsMatched(params url.Values, routeMatcher *gitRouteMatcher) bool {
	if routeMatcher.Params == nil {
		return true
	}
	for _, param := range routeMatcher.Params {
		if _, ok := params[param]; ok {
			return true
		}
	}
	return false
}

func (g *GitServer) absoluteRepoPath(relativePath string) (string, error) {
	if !strings.HasSuffix(relativePath, ".git") {
		relativePath += ".git"
	}
	path := fmt.Sprintf("%s/%s", g.rootDir, relativePath)
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if strings.Contains(path, "..") {
		return "", errors.New("invalid repo path")
	}
	return absolutePath, nil
}

func (g *GitServer) getInfoRefs(route *gitRoute, w http.ResponseWriter, r *http.Request) {
	repo, err := g.absoluteRepoPath(route.RepoPath)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	if !repoExists(repo) {
		cmd := gitCommand{Args: []string{"init", "--bare", repo}}
		_, err := cmd.Run(true)
		if err != nil {
			w.WriteHeader(404)
			return
		}
	}
	serviceName := getServiceName(r)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/x-git-"+serviceName+"-advertisement")
	str := "# service=git-" + serviceName
	fmt.Fprintf(w, "%.4x%s\n", len(str)+5, str)
	fmt.Fprintf(w, "0000")
	writeGitToHTTP(w, gitCommand{Args: []string{serviceName, "--stateless-rpc", "--advertise-refs", repo}})
}

func (g *GitServer) uploadPack(route *gitRoute, w http.ResponseWriter, r *http.Request) {
	repo, err := g.absoluteRepoPath(route.RepoPath)
	if err != nil {
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(404)
		panic(err)
	}
	writeGitToHTTP(w, gitCommand{ProcInput: bytes.NewReader(requestBody), Args: []string{"upload-pack", "--stateless-rpc", repo}})
}

func (g *GitServer) receivePack(route *gitRoute, w http.ResponseWriter, r *http.Request) {
	repo, err := g.absoluteRepoPath(route.RepoPath)
	if err != nil {
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(404)
		panic(err)
	}
	writeGitToHTTP(w, gitCommand{ProcInput: bytes.NewReader(requestBody), Args: []string{"receive-pack", "--stateless-rpc", repo}})
}

func (g *GitServer) goGettable(route *gitRoute, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	url := fmt.Sprintf("%s%s", r.Host, r.URL.Path)
	fmt.Fprintf(w, `<html><head><meta name="go-import" content="%s git https://%s"></head><body>go get %s</body></html>`, url, url, url)
}

func repoExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func getServiceName(r *http.Request) string {
	if len(r.Form["service"]) > 0 {
		return strings.Replace(r.Form["service"][0], "git-", "", 1)
	}
	return ""
}
