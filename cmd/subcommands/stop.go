package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	"io/ioutil"
)

// Stop implements the `stop` subcommand.
var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString("json")
		if err != nil {
			return err
		}
		if path == "" {
			return errors.New("--json not provided")
		}
		// Read the file
		var spec gerrittest.ServiceSpec
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &spec); err != nil {
			return err
		}

		// Terminate the container.
		docker, err := dockertest.NewClient()
		if err != nil {
			return err
		}
		return docker.RemoveContainer(context.Background(), spec.Container)
	},
}

func init() {
	Stop.Flags().String(
		"json", "",
		"The json file containing the container to stop.")
}
