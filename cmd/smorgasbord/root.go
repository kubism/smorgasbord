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
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rootCmd := &cobra.Command{
		Use:   "smorgasbord action [flags]",
		Short: "Smorgasbord is a self-service tool for wireguard users.",
		Long:  `Smorgasbord is a self-service tool for wireguard users.`,
	}
	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)
	// Add all sub-commands
	versionCmd := newVersionCmd(os.Stdout)
	rootCmd.AddCommand(versionCmd)
	serverCmd := newServerCmd(os.Stdout)
	rootCmd.AddCommand(serverCmd)
	setupCmd := newSetupCmd(os.Stdout)
	rootCmd.AddCommand(setupCmd)
	// Make sure to cancel the context if a signal was received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Warn().Str("signal", sig.String()).Msg("received signal")
		cancel()
	}()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error().Err(err).Msg("command failed")
		os.Exit(1)
	}
}
