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
	"fmt"
	"net"
	"net/http"
	"testing"

	_ "github.com/kubism/smorgasbord/internal/flags"
	"github.com/kubism/smorgasbord/pkg/auth"
	"github.com/kubism/smorgasbord/pkg/testutil"
	"github.com/kubism/smorgasbord/pkg/util"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dex       *testutil.Dex
	server    *http.Server
	serverLis net.Listener
	handler   *auth.Handler
	client    *auth.Client
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "auth")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	serverPort, err := util.GetFreePort()
	Expect(err).ToNot(HaveOccurred())
	serverAddr := fmt.Sprintf("127.0.0.1:%d", serverPort)
	redirectURL := fmt.Sprintf("http://%s/auth/callback", serverAddr)
	dex, err = testutil.NewDex(redirectURL)
	Expect(err).ToNot(HaveOccurred())
	Expect(dex).ToNot(BeNil())
	config := &auth.HandlerConfig{
		ClientID:           testutil.DexClientID,
		ClientSecret:       testutil.DexClientSecret,
		IssuerURL:          dex.GetIssuerURL(),
		AuthCodeURLMutator: dex.GetAuthCodeURLMutator(),
		RedirectURL:        redirectURL,
		Nonce:              "test",
		OfflineAsScope:     false,
	}
	handler, err = auth.NewHandler(config)
	Expect(err).ToNot(HaveOccurred())
	Expect(handler).ToNot(BeNil())
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(cors.Default())
	auth.Register(engine, handler)
	server = &http.Server{Addr: serverAddr, Handler: engine}
	serverLis, err = net.Listen("tcp", serverAddr)
	Expect(err).ToNot(HaveOccurred())
	go func() {
		if err := server.Serve(serverLis); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	client = auth.NewClient(fmt.Sprintf("http://%s", serverAddr))
	Expect(err).ToNot(HaveOccurred())
	Expect(client).ToNot(BeNil())
	close(done)
}, 240)

var _ = AfterSuite(func() {
	if dex != nil {
		_ = dex.Close()
	}
	if server != nil {
		_ = server.Close()
	}
	if serverLis != nil {
		_ = serverLis.Close()
	}
	if client != nil {
		_ = client.Close()
	}
})
