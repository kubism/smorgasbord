package auth

import (
	"fmt"
	"net"
	"net/http"

	"github.com/kubism/smorgasbord/pkg/util"

	"github.com/gin-gonic/gin"
)

type Client struct {
	baseURL     string
	callbackURL string
	client      *http.Client
	server      *http.Server
	serverLis   net.Listener
	received    chan string
	token       string
}

func NewClient(baseURL string) (*Client, error) {
	received := make(chan string, 1)
	engine := gin.New()
	engine.GET("/callback", func(c *gin.Context) {
		token := c.Query(QueryTokenKey)
		if token == "" {
			c.String(http.StatusBadRequest, "Did not receive token")
			return
		}
		c.String(http.StatusOK, "Successfully logged in navigate to terminal.")
		received <- token
	})
	port, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: engine,
	}
	serverLis, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := server.Serve(serverLis); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return &Client{
		baseURL:     baseURL,
		callbackURL: fmt.Sprintf("http://%s/callback", server.Addr),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		server:    server,
		serverLis: serverLis,
		received:  received,
	}, nil
}

func (c *Client) Close() error {
	close(c.received)
	if c.server != nil {
		if err := c.server.Close(); err != nil {
			return err
		}
	}
	if err := c.serverLis.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetAuthCodeURL() (string, error) {
	res, err := c.client.Get(fmt.Sprintf("%s/auth/login?callback=%s", c.baseURL, c.callbackURL))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusSeeOther {
		return "", fmt.Errorf("unexpected status code received: %d", res.StatusCode)
	}
	u, err := res.Location()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (c *Client) WaitUntilReceived() {
	c.token = <-c.received
}
