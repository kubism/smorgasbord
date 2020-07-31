package testutil

import (
	"testing"

	_ "github.com/kubism/smorgasbord/internal/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "testutil")
}
