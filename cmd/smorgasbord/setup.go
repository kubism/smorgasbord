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
	"fmt"
	"io"
	"os"

	cfg "github.com/kubism/smorgasbord/pkg/config"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newSetupCmd(out io.Writer) *cobra.Command {
	var config string
	var baseURL string

	cmd := &cobra.Command{
		Use:           "setup",
		Short:         "Configures the environment for subsequent commands.",
		Long:          `Configures the environment for subsequent commands.`,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Well, the output is meant to be consumed by the user, so let's
			// properly setup the log output, both global and locally
			zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			log := zerolog.New(zerolog.ConsoleWriter{Out: out}).With().Timestamp().Logger()
			// Make sure to expand env for config, e.g. $HOME in default
			config = os.ExpandEnv(config)
			if baseURL == "" {
				return fmt.Errorf("Please provide the --base-url flag to setup your configuration")
			}
			c, err := cfg.FromFile(config)
			if os.IsNotExist(err) {
				c = &cfg.Config{Path: config}
			} else if err != nil {
				return fmt.Errorf("Failed to load configuration: %w", err)
			}
			c.BaseURL = baseURL
			err = c.Save()
			if err != nil {
				return fmt.Errorf("Failed to save configuration: %w", err)
			}
			log.Info().Str("config", config).Msg("Wrote changes to configuration")
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "$HOME/.smorgasbord", "Configuration which is updated by the setup command.")
	flags.StringVarP(&baseURL, "base-url", "u", "", "Defines which smorgasbord server subsequent commands will connect to.")

	return cmd
}
