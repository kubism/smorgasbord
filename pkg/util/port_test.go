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