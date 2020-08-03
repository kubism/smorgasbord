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
	"io"
	"net"
	"net/http"
	"time"

	"github.com/kubism/smorgasbord/pkg/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func newServerCmd(out io.Writer) *cobra.Command {
	var (
		addr                string
		clientID            string
		clientSecret        string
		issuerURL           string
		redirectURL         string
		authCodeURLAppendix string
		nonce               string
		debug               bool
	)

	authCodeURLMutator := func(url string) string {
		if authCodeURLAppendix != "" {
			return url + authCodeURLAppendix
		}
		return url
	}

	cmd := &cobra.Command{
		Use:           "server",
		Short:         "Starts the smorgasbord server.",
		Long:          `Starts the smorgasbord server.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			// Setup logger
			log := zerolog.New(out).With().Timestamp().Logger()
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			// Setup auth.Handler which handles the OIDC flows
			config := &auth.HandlerConfig{
				ClientID:           clientID,
				ClientSecret:       clientSecret,
				IssuerURL:          issuerURL,
				AuthCodeURLMutator: authCodeURLMutator,
				RedirectURL:        redirectURL,
				Nonce:              nonce,
				OfflineAsScope:     false,
			}
			handler, err := auth.NewHandler(config)
			if err != nil {
				return err
			}
			// Setup gin with logger
			if !debug {
				gin.SetMode(gin.ReleaseMode)
			}
			engine := gin.New()
			engine.Use(gin.Recovery())
			engine.Use(cors.Default())
			engine.Use(logger.SetLogger(logger.Config{
				Logger: &log,
				UTC:    true,
			}))
			auth.Register(engine, handler)
			// Create the http server and listen on address
			server := &http.Server{Addr: addr, Handler: engine}
			log.Info().Str("addr", addr).Msg("starting listener")
			serverLis, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			defer func() {
				_ = serverLis.Close()
			}()
			go func() { // Start listening
				log.Info().Msg("server starting")
				if err := server.Serve(serverLis); err != http.ErrServerClosed {
					panic(err)
				}
				log.Info().Msg("server shutdown")
			}()
			<-ctx.Done()
			log.Info().Msg("context cancelled or timeout, shutting down server")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return server.Shutdown(ctx)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&addr, "addr", "a", "0.0.0.0:8080", "Which address the server will listen on.")
	flags.StringVarP(&clientID, "client-id", "c", "", "OIDC/OAuth2 client ID used for OIDC flow.")
	flags.StringVarP(&clientSecret, "client-secret", "s", "", "OIDC/OAuth2 client secret used for OIDC flow.")
	flags.StringVarP(&issuerURL, "issuer-url", "i", "", "Issuer URL for OIDC flow, e.g. auth code retrieval.")
	flags.StringVarP(&redirectURL, "redirect-url", "r", "", "Public redirect URL pointing to the callback of the server as configured for the client.")
	flags.StringVarP(&authCodeURLAppendix, "auth-code-url-appendix", "x", "", "Some OIDC providers will not return the full auth code URL, this flag can be used to append to the URL (e.g. for dex connector selection).")
	flags.StringVarP(&nonce, "nonce", "n", "", "Nonce used to hash state during redirect flow (keep it secret).")
	flags.BoolVar(&debug, "debug", false, "Whether to use debug mode for the server and log.")

	return cmd
}
