// Taken from dex, therefore their license applies.
// Removed functionality to required for tests.
package testutil

import (
	"github.com/dexidp/dex/pkg/log"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/memory"
)

// Config is the config format for the main application.
type Config struct {
	Issuer    string    `json:"issuer"`
	Storage   Storage   `json:"storage"`
	Web       Web       `json:"web"`
	Telemetry Telemetry `json:"telemetry"`
	OAuth2    OAuth2    `json:"oauth2"`
	Expiry    Expiry    `json:"expiry"`
	Logger    Logger    `json:"logger"`

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

// Telemetry is the config format for telemetry including the HTTP server config.
type Telemetry struct {
	HTTP string `json:"http"`
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

var storages = map[string]func() StorageConfig{
	"memory": func() StorageConfig { return new(memory.Config) },
}

// Expiry holds configuration for the validity period of components.
type Expiry struct {
	// SigningKeys defines the duration of time after which the SigningKeys will be rotated.
	SigningKeys string `json:"signingKeys"`

	// IdTokens defines the duration of time for which the IdTokens will be valid.
	IDTokens string `json:"idTokens"`

	// AuthRequests defines the duration of time for which the AuthRequests will be valid.
	AuthRequests string `json:"authRequests"`
}

// Logger holds configuration required to customize logging for dex.
type Logger struct {
	// Level sets logging level severity.
	Level string `json:"level"`

	// Format specifies the format to be used for logging.
	Format string `json:"format"`
}
