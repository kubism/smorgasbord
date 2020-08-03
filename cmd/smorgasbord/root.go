package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "smorgasbord action [flags]",
	Short:        "Smorgasbord is a self-service tool for wireguard users.",
	Long:         `Smorgasbord is a self-service tool for wireguard users.`,
	SilenceUsage: true,
}

func init() {
	// Setup global flags
	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
