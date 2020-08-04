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
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	expected = []byte(`{ "baseURL": "b", "token": "t" }`)
)

var _ = Describe("Config", func() {
	It("can be loaded from raw and saved", func() {
		c, err := FromRaw(expected)
		Expect(err).ToNot(HaveOccurred())
		Expect(c).ToNot(BeNil())
		Expect(c.Save()).ToNot(Succeed())
		Expect(c.SaveTo(filepath.Join(tmpDir, "config1"))).To(Succeed())
	})
	It("can be loaded from file and saved", func() {
		path := filepath.Join(tmpDir, "config2")
		Expect(ioutil.WriteFile(path, expected, 0644)).To(Succeed())
		c, err := FromFile(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(c).ToNot(BeNil())
		Expect(c.Save()).To(Succeed())
		Expect(c.SaveTo(filepath.Join(tmpDir, "config3"))).To(Succeed())
	})
	It("fails for raw malformed json", func() {
		c, err := FromRaw([]byte(`{]`))
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})
	It("fails for file with malformed json", func() {
		path := filepath.Join(tmpDir, "config4")
		Expect(ioutil.WriteFile(path, []byte(`{]`), 0644)).To(Succeed())
		c, err := FromFile(path)
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})
	It("fails if file does not exist", func() {
		path := filepath.Join(tmpDir, "configdoesnotexist")
		c, err := FromFile(path)
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})
})
