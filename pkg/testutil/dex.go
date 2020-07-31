package testutil

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kubism/smorgasbord/internal/flags"

	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/memory"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	DexClientSecret     = "ZXhhbXBsZS1hcHAtc2VjcmV0"
	DexUserEmail        = "test@kubism.io"
	DexUserPassword     = "password"
	dexUserPasswordHash = "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
)

type DebugLog = func(string, ...interface{})

type dexLog struct {
	log DebugLog
}

func (l *dexLog) Debug(args ...interface{}) {
	l.log("DEBUG: " + fmt.Sprint(args...))
}

func (l *dexLog) Info(args ...interface{}) {
	l.log("INFO: " + fmt.Sprint(args...))
}

func (l *dexLog) Warn(args ...interface{}) {
	l.log("WARN: " + fmt.Sprint(args...))
}

func (l *dexLog) Error(args ...interface{}) {
	l.log("ERROR: " + fmt.Sprint(args...))
}

func (l *dexLog) Debugf(format string, args ...interface{}) {
	l.log("DEBUG: " + fmt.Sprintf(format, args...))
}

func (l *dexLog) Infof(format string, args ...interface{}) {
	l.log("INFO: " + fmt.Sprintf(format, args...))
}

func (l *dexLog) Warnf(format string, args ...interface{}) {
	l.log("WARN " + fmt.Sprintf(format, args...))

}

func (l *dexLog) Errorf(format string, args ...interface{}) {
	l.log("ERROR: " + fmt.Sprintf(format, args...))
}

type Dex struct {
	server *http.Server
}

func NewDex(redirectURI string) (*Dex, error) {
	port := GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	c := Config{
		Issuer: fmt.Sprintf("http://%s/dex", addr),
		Storage: Storage{
			Type:   "memory",
			Config: &memory.Config{},
		},
		Web: Web{
			AllowedOrigins: []string{"*"},
			HTTP:           addr,
		},
		Frontend: server.WebConfig{
			Dir: flags.DexWebDir,
		},
		OAuth2: OAuth2{
			SkipApprovalScreen: true,
		},
		StaticClients: []storage.Client{
			{
				ID:           "smorgasbord",
				RedirectURIs: []string{redirectURI},
				Name:         "Smorgasbord",
				Secret:       DexClientSecret,
			},
		},
		StaticConnectors: []storage.Connector{
			{
				ID:   "mock",
				Name: "Mock",
				Type: "mockCallback",
			},
			{
				ID:   server.LocalConnector,
				Name: "Email",
				Type: server.LocalConnector,
			},
		},
		StaticPasswords: []storage.Password{
			{
				Email:    DexUserEmail,
				Hash:     []byte(dexUserPasswordHash),
				Username: "test",
				UserID:   "08a8684b-db88-4b73-90a9-3cd1661f5466",
			},
		},
	}
	log := &dexLog{func(string, ...interface{}) {}}
	s, _ := c.Storage.Config.Open(log)
	s = storage.WithStaticClients(s, c.StaticClients)
	s = storage.WithStaticPasswords(s, c.StaticPasswords, log)
	s = storage.WithStaticConnectors(s, c.StaticConnectors)
	serverConfig := server.Config{
		SupportedResponseTypes: c.OAuth2.ResponseTypes,
		SkipApprovalScreen:     c.OAuth2.SkipApprovalScreen,
		AlwaysShowLoginScreen:  c.OAuth2.AlwaysShowLoginScreen,
		PasswordConnector:      c.OAuth2.PasswordConnector,
		AllowedOrigins:         c.Web.AllowedOrigins,
		Issuer:                 c.Issuer,
		Storage:                s,
		Web:                    c.Frontend,
		Logger:                 log,
		Now:                    func() time.Time { return time.Now().UTC() },
		PrometheusRegistry:     prometheus.NewRegistry(),
	}
	handler, err := server.NewServer(context.Background(), serverConfig)
	if err != nil {
		return nil, err
	}
	httpServer := &http.Server{Addr: c.Web.HTTP, Handler: handler}
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			panic(err) // unexpected error. port in use?
		}
	}()
	return &Dex{httpServer}, nil
}

func (d *Dex) Close() error {
	return d.server.Close()
}
