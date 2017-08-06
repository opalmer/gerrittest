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
	Start.Flags().String(
		"json", "",
		"The location to write information about the service to. Any "+
			"existing content will be overwritten.")
	Start.Flags().String(
		"image", "opalmer/gerrittest:2.14.2",
		"The Docker image to spin up Gerrit.")
	Start.Flags().Uint16(
		"port-http", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
	Start.Flags().Uint16(
		"port-ssh", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
}
