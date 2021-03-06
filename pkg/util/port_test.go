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

package util

import (
	"net"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetFreePort", func() {
	It("gets free port", func() {
		port, err := GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		Expect(port).To(BeNumerically(">", 0))
		l, err := net.Listen("tcp", "localhost"+":"+strconv.Itoa(port))
		Expect(err).ToNot(HaveOccurred())
		defer l.Close()
	})
})
