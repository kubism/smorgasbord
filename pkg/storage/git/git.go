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
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/kubism/smorgasbord/pkg/storage"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitstorage "github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
)

const stateName = "smorgasbord.json"

type state = map[string][]storage.Entry

type gitStorage struct {
	auth   transport.AuthMethod
	fs     billy.Filesystem
	storer gitstorage.Storer
	repo   *git.Repository
}

func NewStorage(repositoryURL string, auth transport.AuthMethod) (storage.Storage, error) {
	s := &gitStorage{
		auth:   auth,
		fs:     memfs.New(),
		storer: memory.NewStorage(),
	}
	return s, s.clone(repositoryURL)
}

func (s *gitStorage) Add(id, publicKey string) error {
	return nil
}

func (s *gitStorage) Delete(id, publicKey string) error {
	return nil
}

func (s *gitStorage) List(id string) ([]storage.Entry, error) {
	st, err := s.load()
	if err != nil {
		return nil, err
	}
	entries, ok := st[id]
	if !ok || entries == nil {
		return []storage.Entry{}, nil
	}
	return entries, nil
}

func (s *gitStorage) Save() error {
	return nil
}

func (s *gitStorage) Close() error {
	return nil
}

func (s *gitStorage) clone(url string) error {
	var err error
	s.repo, err = git.Clone(s.storer, s.fs, &git.CloneOptions{
		URL:   url,
		Auth:  s.auth,
		Depth: 5,
	})
	return err
}

func (s *gitStorage) load() (state, error) {
	if err := s.pull(); err != nil {
		return nil, err
	}
	f, err := s.fs.Open(stateName)
	if os.IsNotExist(err) {
		return state{}, nil
	} else if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	st := state{}
	err = json.Unmarshal(data, &st)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (s *gitStorage) pull() error {
	w, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}
