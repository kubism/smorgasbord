package auth

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	_ "github.com/kubism/smorgasbord/internal/flags"
	"github.com/kubism/smorgasbord/pkg/testutil"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dex       *testutil.Dex
	server    *http.Server
	serverLis net.Listener
	handler   *Handler
	client    *Client
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "auth")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	serverPort := testutil.GetFreePort()
	serverAddr := fmt.Sprintf("127.0.0.1:%d", serverPort)
	redirectURL := fmt.Sprintf("http://%s/auth/callback", serverAddr)
	dex, err = testutil.NewDex(redirectURL)
	Expect(err).ToNot(HaveOccurred())
	Expect(dex).ToNot(BeNil())
	config := &HandlerConfig{
		ClientID:           testutil.DexClientID,
		ClientSecret:       testutil.DexClientSecret,
		IssuerURL:          dex.GetIssuerURL(),
		AuthCodeURLMutator: dex.GetAuthCodeURLMutator(),
		RedirectURL:        redirectURL,
		Nonce:              "test",
		OfflineAsScope:     true,
	}
	handler, err = NewHandler(config)
	Expect(err).ToNot(HaveOccurred())
	Expect(handler).ToNot(BeNil())
	engine := gin.Default()
	engine.Use(cors.Default())
	Register(engine, handler)
	server = &http.Server{Addr: serverAddr, Handler: engine}
	serverLis, err = net.Listen("tcp", serverAddr)
	Expect(err).ToNot(HaveOccurred())
	go func() {
		if err := server.Serve(serverLis); err != http.ErrServerClosed {
			panic(err) // unexpected error. port in use?
		}
	}()
	client, err = NewClient(fmt.Sprintf("http://%s", serverAddr))
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
