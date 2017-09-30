package cmd

import (
	"errors"

	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
)

// Stop implements the `stop` subcommand.
var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Responsible for stopping an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := getString(cmd, "json")
		if path == "" {
			return errors.New("--json not provided")
		}

		instance, err := gerrittest.NewFromJSON(path)
		if err != nil {
			return err
		}
		return instance.Destroy()
	},
}

func init() {
	addCommonFlags(Stop)
}
