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
	"io"
	"os"
	"time"

	"github.com/kubism/smorgasbord/pkg/auth"
	cfg "github.com/kubism/smorgasbord/pkg/config"

	"github.com/pkg/browser"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var openURL = browser.OpenURL

func newLoginCmd(out io.Writer) *cobra.Command {
	var (
		config string
	)

	cmd := &cobra.Command{
		Use:           "setup",
		Short:         "Configures the environment for subsequent commands.",
		Long:          `Configures the environment for subsequent commands.`,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			// Well, the output is meant to be consumed by the user, so let's
			// properly setup the log output, both global and locally
			zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			log := zerolog.New(zerolog.ConsoleWriter{Out: out}).With().Timestamp().Logger()
			// Make sure to expand env for config, e.g. $HOME in default
			config = os.ExpandEnv(config)
			c, err := cfg.FromFile(config)
			if os.IsNotExist(err) {
				c = &cfg.Config{Path: config}
			} else if err != nil {
				return fmt.Errorf("Failed to load configuration: %w", err)
			}
			client := auth.NewClient(c.BaseURL)
			if err := client.StartCallbackServer(); err != nil {
				return fmt.Errorf("Failed to start callback server: %w", err)
			}
			defer func() {
				_ = client.StopCallbackServer()
			}()
			authCodeURL, err := client.GetAuthCodeURL()
			if err != nil {
				return fmt.Errorf("Failed to retrieve auth code URL: %w", err)
			}
			if err := openURL(authCodeURL); err != nil {
				return fmt.Errorf("Failed to open auth code URL in browser: %w", err)
			}
			log.Info().Msg("Waiting up to 120 seconds for token response...")
			ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
			defer cancel()
			if err := client.WaitUntilTokenReceived(ctx); err != nil {
				return fmt.Errorf("Failed to receive token: %w", err)
			}
			c.Token = client.GetToken()
			if err := c.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			log.Info().Str("config", config).Msg("Login successful. Writing changes to configuration.")
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "$HOME/.smorgasbord", "Configuration which is to login.")

	return cmd
}
