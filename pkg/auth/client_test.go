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

package auth_test

import (
	"net/http"
	"net/url"
	"time"

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
		Expect(client.StartCallbackServer()).To(Succeed())
		defer func() {
			Expect(client.StopCallbackServer()).To(Succeed())
		}()
		authCodeURL, err := client.GetAuthCodeURL()
		Expect(err).ToNot(HaveOccurred())
		Expect(authCodeURL).ToNot(Equal(""))
		_, err = url.Parse(authCodeURL)
		Expect(err).ToNot(HaveOccurred())
	})
	It("can log user in and retrieve token", func() {
		Expect(client.StartCallbackServer()).To(Succeed())
		authCodeURL, err := client.GetAuthCodeURL()
		Expect(err).ToNot(HaveOccurred())
		simulateUserLoginInBrowser(authCodeURL)
		Expect(client.WaitUntilReceived(5 * time.Second)).To(Succeed())
		Expect(client.GetToken()).ToNot(Equal(""))
		Expect(client.StopCallbackServer()).To(Succeed())
	})
})
