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

package main

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Setup", func() {
	It("works with all flags provided", func() {
		output, err := executeCommandWithContext(context.Background(), newSetupCmd, validSetupArgs()...)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).ToNot(Equal(""))
	})
	It("fails without proper flags", func() {
		_, err := executeCommandWithContext(context.Background(), newSetupCmd)
		Expect(err).To(HaveOccurred())
	})
})
