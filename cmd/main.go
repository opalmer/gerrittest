package main

import (
	"fmt" 
	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/gerrittest/cmd/subcommands/start"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "gerrittest",
	Short: "Command line tool for testing and working with Gerrit in Docker.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		value, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}
		level, err := log.ParseLevel(value)
		if err != nil {
			return err
		}
		log.SetLevel(level)
		return nil
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
	RootCmd.AddCommand(start.Cmd)
}
