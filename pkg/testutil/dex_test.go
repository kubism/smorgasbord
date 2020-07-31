package testutil

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dex", func() {
	It("is created successfully", func() {
		dex, err := NewDex("test")
		Expect(err).ToNot(HaveOccurred())
		defer dex.Close()
		Expect(dex).ToNot(BeNil())
	})
})
