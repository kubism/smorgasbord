package testutil

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LDAP", func() {
	It("is created successfully", func() {
		ldap, err := NewLDAP()
		Expect(err).ToNot(HaveOccurred())
		defer ldap.Close()
		Expect(ldap).ToNot(BeNil())
	})
})
