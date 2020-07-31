package auth

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func simulateUserLoginInBrowser(authCodeURL string) {
	// Dex's mock connector will simply create a token and not require any user
	// input, so really simulating is atm not required.
	c := &http.Client{}
	res, err := c.Get(authCodeURL)
	Expect(err).ToNot(HaveOccurred())
	defer res.Body.Close()
	Expect(res.StatusCode).To(Equal(http.StatusOK))
}

var _ = Describe("Client", func() {
	It("can retrieve auth code URL", func() {
		authCodeURL, err := client.GetAuthCodeURL()
		Expect(err).ToNot(HaveOccurred())
		Expect(authCodeURL).ToNot(Equal(""))
		_, err = url.Parse(authCodeURL)
		Expect(err).ToNot(HaveOccurred())
	})
	It("can log user in and retrieve token", func() {
		authCodeURL, err := client.GetAuthCodeURL()
		Expect(err).ToNot(HaveOccurred())
		simulateUserLoginInBrowser(authCodeURL)
		client.WaitUntilReceived()
		Expect(client.token).ToNot(Equal(""))
	})
})
