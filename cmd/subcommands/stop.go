package cmd

import (
	"errors"

	"github.com/opalmer/gerrittest"
	"github.com/opalmer/logrusutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Stop implements the `stop` subcommand.
var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Responsible for stopping an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup logging
		cfg := logrusutil.NewConfig()
		cfg.Level = getString(cmd, "log-level")
		if err := logrusutil.ConfigureLogger(log.StandardLogger(), cfg); err != nil {
			return err
		}

		path := getString(cmd, "json")
		if path == "" {
			return errors.New("--json not provided")
		}

		gerrit, err := gerrittest.NewFromJSON(getString(cmd, "json"))
		if err != nil {
			return err
		}
		return gerrit.Destroy()
	},
}

func init() {
	Stop.Flags().String(
		"json", "",
		"The json file containing the container to stop.")
	Stop.Flags().String(
		"log-level", "panic",
		"Configures the logging level")
}
