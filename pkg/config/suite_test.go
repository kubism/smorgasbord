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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/kubism/smorgasbord/internal/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	tmpDir string
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/config")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	tmpDir, err = ioutil.TempDir("", "smorgasbord")
	Expect(err).ToNot(HaveOccurred())
	close(done)
}, 240)

var _ = AfterSuite(func() {
	if tmpDir != "" {
		_ = os.RemoveAll(tmpDir)
	}
})
