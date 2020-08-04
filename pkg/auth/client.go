/*
Copyright 2020 Smorgasbord Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/kubism/smorgasbord/pkg/util"

	"github.com/gin-gonic/gin"
)

type Client struct {
	baseURL     string
	callbackURL string
	client      *http.Client
	server      *http.Server
	serverLis   net.Listener
	received    chan string
	token       string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		received: make(chan string, 1),
	}
}

func (c *Client) StartCallbackServer() error {
	// NOTE: the callback server basically consists of three components, which
	// will be cleaned up by StopCallbackServer:
	// server, serverLis and callbackURL
	engine := gin.New()
	engine.GET("/callback", func(g *gin.Context) {
		token := g.Query(QueryTokenKey)
		if token == "" {
			g.String(http.StatusBadRequest, "Did not receive token")
			return
		}
		g.String(http.StatusOK, "Successfully logged in navigate to terminal.")
		c.received <- token
	})
	port, err := util.GetFreePort()
	if err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: engine,
	}
	c.callbackURL = fmt.Sprintf("http://%s/callback", c.server.Addr)
	c.serverLis, err = net.Listen("tcp", c.server.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := c.server.Serve(c.serverLis); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return nil
}

// GetAuthCodeURL will connect to the server and retrieve the URL, which
// the client will be redirected to as part of the OIDC flow.
// Make sure to start the callback server via StartCallbackServer first or
// this function will return an error.
func (c *Client) GetAuthCodeURL() (string, error) {
	if c.callbackURL == "" {
		return "", fmt.Errorf("callback URL not available, make sure to start the callback server first")
	}
	res, err := c.client.Get(fmt.Sprintf("%s/auth/login?callback=%s", c.baseURL, c.callbackURL))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusSeeOther {
		return "", fmt.Errorf("unexpected status code received: %d", res.StatusCode)
	}
	u, err := res.Location()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// WaitUntilTokenReceived will block until either the token was received or
// the context is done. The token can only be received if the callback server
// is running and the user is finishing the OIDC flow with his browser.
func (c *Client) WaitUntilTokenReceived(ctx context.Context) error {
	select {
	case t := <-c.received:
		c.token = t
	case <-ctx.Done():
		return fmt.Errorf("failed to receive token before context done")
	}
	return nil
}

// StopCallbackServer will shutdown the server, which receives the token by
// being the final redirect as part of the OIDC flow.
func (c *Client) StopCallbackServer() error {
	var err error
	c.callbackURL = ""
	if c.server != nil {
		err = c.server.Shutdown(context.Background())
		c.server = nil
	}
	if c.serverLis != nil {
		_ = c.serverLis.Close()
		c.serverLis = nil
	}
	return err
}

// Close will clean all listeners and channels required by the Client.
func (c *Client) Close() error {
	close(c.received)
	return c.StopCallbackServer()
}

// GetToken return the current value of the token. If the token has not been
// retrieved, the value will be an empty string.
func (c *Client) GetToken() string {
	return c.token
}
