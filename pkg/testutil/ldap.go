package testutil

import (
	"github.com/glauth/glauth/pkg/config"
	"github.com/glauth/glauth/pkg/handler"
	"github.com/nmcclain/ldap"
	logging "github.com/op/go-logging"
)

const (
	BaseDN         = "dc=kubism,dc=io"
	Password       = "dogood"
	passwordSHA256 = "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a"
)

type LDAP struct {
	server *ldap.Server
	quit   chan bool
}

func NewLDAP() (*LDAP, error) {
	config := config.Config{
		Backend: config.Backend{
			BaseDN:    BaseDN,
			Datastore: "config",
		},
		Users: []config.User{
			{
				Name:         "foo",
				UnixID:       5001,
				PrimaryGroup: 5501,
				PassSHA256:   passwordSHA256,
			},
			{
				Name:         "bar",
				UnixID:       5002,
				PrimaryGroup: 5502,
				PassSHA256:   passwordSHA256,
			},
		},
		Groups: []config.Group{
			{
				Name:   "admins",
				UnixID: 5501,
			},
			{
				Name:   "users",
				UnixID: 5502,
			},
		},
	}
	log := logging.MustGetLogger("ldap")
	h := handler.NewConfigHandler(log, &config, nil)
	s := ldap.NewServer()
	s.EnforceLDAP = true
	s.BindFunc("", h)
	s.SearchFunc("", h)
	s.CloseFunc("", h)
	quit := make(chan bool)
	s.QuitChannel(quit)
	go func() {
		err := s.ListenAndServe("0.0.0.0:10389")
		if err != nil {
			panic(err)
		}
	}()
	return &LDAP{
		server: s,
		quit:   quit,
	}, nil
}

func (l *LDAP) Close() {
	l.quit <- true
}
