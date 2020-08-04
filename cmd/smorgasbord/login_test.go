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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testOpenURL = func(url string) error {
	c := &http.Client{}
	res, err := c.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Unexpected status code, expected 200, got %d with body: %s", res.StatusCode, data)
	}
	return nil
}

var _ = Describe("Login", func() {
	It("works with if properly setup", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_, err := executeCommandWithContext(ctx, newServerCmd, validServerArgs()...)
			Expect(err).ToNot(HaveOccurred())
		}()
		_, err := executeCommandWithContext(context.Background(), newSetupCmd, validSetupArgs()...)
		Expect(err).ToNot(HaveOccurred())
		Expect(waitUntilServerReady()).To(Succeed())
		openURL = testOpenURL
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		output, err := executeCommandWithContext(ctx, newLoginCmd, validLoginArgs()...)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("Login successful"))
	})
})
