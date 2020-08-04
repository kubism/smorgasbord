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

package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	_ "github.com/kubism/smorgasbord/internal/flags"
	"github.com/kubism/smorgasbord/pkg/storage"
	"github.com/kubism/smorgasbord/pkg/testutil"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const testID = "test@test.com"

var (
	gitServer *testutil.GitServer
	gitS      storage.Storage
	tmpDir    string
)

func TestGitStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/storage/git")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	tmpDir, err = ioutil.TempDir("", "smorgasbord")
	Expect(err).ToNot(HaveOccurred())
	gitServer, err = testutil.NewGitServer(tmpDir)
	Expect(err).ToNot(HaveOccurred())
	// Let's setup the repository for first use
	fs := memfs.New()
	r, err := git.Init(memory.NewStorage(), fs)
	Expect(err).ToNot(HaveOccurred())
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{getRemoteURL()},
	})
	Expect(err).ToNot(HaveOccurred())
	f, err := fs.Create(stateName)
	Expect(err).ToNot(HaveOccurred())
	_, err = io.WriteString(f, `{ "test@test.com": [{ "publicKey": "...", "allowedIP": "0.0.0.0/0" }] }`)
	Expect(err).ToNot(HaveOccurred())
	Expect(f.Close()).To(Succeed())
	w, err := r.Worktree()
	Expect(err).ToNot(HaveOccurred())
	_, err = w.Add(stateName)
	Expect(err).ToNot(HaveOccurred())
	_, err = w.Commit("first commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(r.Push(&git.PushOptions{
		RemoteName: "origin",
	})).To(Succeed())
	// Lastly setup gitStorage
	gitS, err = NewStorage(getRemoteURL(), nil)
	Expect(err).ToNot(HaveOccurred())
	close(done)
}, 240)

var _ = AfterSuite(func() {
	if gitServer != nil {
		_ = gitServer.Close()
	}
	if tmpDir != "" {
		_ = os.RemoveAll(tmpDir)
	}
})

func getRemoteURL() string {
	return fmt.Sprintf("http://%s/test.git", gitServer.GetAddr())
}
