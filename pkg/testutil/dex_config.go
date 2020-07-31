// Taken from dex, therefore their license applies.
// Removed functionality to required for tests.
package testutil

import (
	"github.com/dexidp/dex/pkg/log"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
)

// Config is the config format for the main application.
type Config struct {
	Issuer  string  `json:"issuer"`
	Storage Storage `json:"storage"`
	Web     Web     `json:"web"`
	OAuth2  OAuth2  `json:"oauth2"`

	Frontend server.WebConfig `json:"frontend"`

	// StaticConnectors are user defined connectors specified in the ConfigMap
	// Write operations, like updating a connector, will fail.
	StaticConnectors []storage.Connector `json:"connectors"`

	// StaticClients cause the server to use this list of clients rather than
	// querying the storage. Write operations, like creating a client, will fail.
	StaticClients []storage.Client `json:"staticClients"`

	// StaticPasswords cause the server use this list of passwords rather than
	// querying the storage. Cannot be specified without enabling a passwords
	// database.
	StaticPasswords []storage.Password `json:"staticPasswords"`
}

// OAuth2 describes enabled OAuth2 extensions.
type OAuth2 struct {
	ResponseTypes []string `json:"responseTypes"`
	// If specified, do not prompt the user to approve client authorization. The
	// act of logging in implies authorization.
	SkipApprovalScreen bool `json:"skipApprovalScreen"`
	// If specified, show the connector selection screen even if there's only one
	AlwaysShowLoginScreen bool `json:"alwaysShowLoginScreen"`
	// This is the connector that can be used for password grant
	PasswordConnector string `json:"passwordConnector"`
}

// Web is the config format for the HTTP server.
type Web struct {
	HTTP           string   `json:"http"`
	HTTPS          string   `json:"https"`
	TLSCert        string   `json:"tlsCert"`
	TLSKey         string   `json:"tlsKey"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

// Storage holds app's storage configuration.
type Storage struct {
	Type   string        `json:"type"`
	Config StorageConfig `json:"config"`
}

// StorageConfig is a configuration that can create a storage.
type StorageConfig interface {
	Open(logger log.Logger) (storage.Storage, error)
}
