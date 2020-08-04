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

package testutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitServer", func() {
	It("is created successfully and repo is functional", func() {
		dir, err := ioutil.TempDir("", "smorgasbord")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)
		gs, err := NewGitServer(dir)
		Expect(err).ToNot(HaveOccurred())
		Expect(gs).ToNot(BeNil())
		defer gs.Close()
		fs := memfs.New()
		r, err := git.Init(memory.NewStorage(), fs)
		Expect(err).ToNot(HaveOccurred())
		_, err = r.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{fmt.Sprintf("http://%s/foo.git", gs.GetAddr())},
		})
		Expect(err).ToNot(HaveOccurred())
		f, err := fs.Create("bar")
		Expect(err).ToNot(HaveOccurred())
		_, err = io.WriteString(f, "hello world")
		Expect(err).ToNot(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		w, err := r.Worktree()
		Expect(err).ToNot(HaveOccurred())
		_, err = w.Add("bar")
		Expect(err).ToNot(HaveOccurred())
		_, err = w.Commit("bar commit", &git.CommitOptions{
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
	})
})
