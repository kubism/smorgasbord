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
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// Setup global flags

}

func main() {
	rootCmd := &cobra.Command{
		Use:          "smorgasbord action [flags]",
		Short:        "Smorgasbord is a self-service tool for wireguard users.",
		Long:         `Smorgasbord is a self-service tool for wireguard users.`,
		SilenceUsage: true,
	}
	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)
	versionCmd := newVersionCmd(os.Stdout)
	rootCmd.AddCommand(versionCmd)
	serverCmd := newServerCmd(os.Stdout)
	rootCmd.AddCommand(serverCmd)
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
