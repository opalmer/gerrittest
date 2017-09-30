package main

import (
	"fmt"
	"os"

	"github.com/opalmer/gerrittest/cmd/subcommands"
	"github.com/opalmer/logrusutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmd is the main command line tool. Subcommands are implemented
// in the subcommands/ folder.
var RootCmd = &cobra.Command{
	Use:   "gerrittest",
	Short: "Command line tool for testing and working with Gerrit in Docker.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		value, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}
		cfg := logrusutil.NewConfig()
		cfg.Level = value
		return logrusutil.ConfigureLogger(log.StandardLogger(), cfg)

	},
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().String(
		"log-level", "warning",
		"Configures the global logging level.")
	RootCmd.AddCommand(cmd.Start)
	RootCmd.AddCommand(cmd.Stop)
	RootCmd.AddCommand(cmd.GetSSHCommand)
}
