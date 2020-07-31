package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type State struct {
	Callback string `form:"callback" json:"callback,omitempty"`
}

type ExtraClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type HandlerConfig struct {
	IssuerURL          string
	OfflineAsScope     bool
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	Nonce              string
	AuthCodeURLMutator func(string) string
}

type Handler struct {
	httpClient *http.Client
	verifier   *oidc.IDTokenVerifier
	provider   *oidc.Provider
	config     *HandlerConfig
}

func NewHandler(config *HandlerConfig) (*Handler, error) {
	var err error
	h := &Handler{
		config:     config,
		httpClient: http.DefaultClient,
	}
	ctx := oidc.ClientContext(context.Background(), h.httpClient)
	h.provider, err = oidc.NewProvider(ctx, h.config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider %q: %v", h.config.IssuerURL, err)
	}
	// What scopes does a provider support?
	var scopes struct {
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		Supported []string `json:"scopes_supported"`
	}
	if err := h.provider.Claims(&scopes); err != nil {
		return nil, fmt.Errorf("failed to parse provider scopes_supported: %v", err)
	}
	if len(scopes.Supported) == 0 {
		// `scopes_supported` is a "RECOMMENDED" discovery claim, not a required
		// one. If missing, assume that the provider follows the spec and has
		// an "offline_access" scope.
		h.config.OfflineAsScope = true
	} else {
		// See if scopes_supported has the "offline_access" scope.
		h.config.OfflineAsScope = func() bool {
			for _, scope := range scopes.Supported {
				if scope == oidc.ScopeOfflineAccess {
					return true
				}
			}
			return false
		}()
	}

	h.verifier = h.provider.Verifier(&oidc.Config{ClientID: h.config.ClientID})
	return h, nil
}

func (h *Handler) GetAuthCodeURL(state *State) (string, error) {
	scopes := []string{oidc.ScopeOpenID, "profile", "email"}
	encoded, err := encode(state)
	if err != nil {
		return "", err
	}
	nonce := hashString(encoded + h.config.Nonce)
	// Construct authCodeURL
	authCodeURL := ""
	if h.config.OfflineAsScope {
		scopes = append(scopes, oidc.ScopeOfflineAccess)
		authCodeURL = h.getOauth2Config(scopes).AuthCodeURL(encoded, oidc.Nonce(nonce))
	} else {
		authCodeURL = h.getOauth2Config(scopes).AuthCodeURL(encoded, oidc.Nonce(nonce), oauth2.AccessTypeOffline)
	}
	if h.config.AuthCodeURLMutator != nil {
		authCodeURL = h.config.AuthCodeURLMutator(authCodeURL)
	}
	return authCodeURL, nil
}

func (h *Handler) Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	t := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour),
	}
	return h.getOauth2Config(nil).TokenSource(h.clientContext(ctx), t).Token()
}

func (h *Handler) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return h.getOauth2Config(nil).Exchange(h.clientContext(ctx), code)
}

func (h *Handler) Verify(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}
	return h.verifier.Verify(h.clientContext(ctx), rawIDToken)
}

func (h *Handler) VerifyStateAndClaims(ctx context.Context, token *oauth2.Token, encoded string) (*State, *ExtraClaims, error) {
	idToken, err := h.Verify(ctx, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify ID token: %v", err)
	}
	if idToken.Nonce != hashString(encoded+h.config.Nonce) {
		return nil, nil, fmt.Errorf("invalid id_token nonce")
	}

	state := &State{}
	err = decode(encoded, state)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode state")
	}

	claims := &ExtraClaims{}
	if err = idToken.Claims(claims); err != nil {
		return nil, nil, fmt.Errorf("claims can not be unmarshalled: %v", err)
	}
	if !claims.EmailVerified {
		return nil, nil, fmt.Errorf("email not verified")
	}
	return state, claims, nil
}

func (h *Handler) getOauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.config.ClientID,
		ClientSecret: h.config.ClientSecret,
		Endpoint:     h.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  h.config.RedirectURL,
	}
}

func (h *Handler) clientContext(ctx context.Context) context.Context {
	return oidc.ClientContext(ctx, h.httpClient)
}

func decode(encoded string, obj interface{}) error {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("error decoding: %v", err)
	}
	return json.Unmarshal(data, obj)
}

func encode(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("error marshalling: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	return encoded, nil
}

func hashString(input string) string {
	sha256Bytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sha256Bytes[:])
}
