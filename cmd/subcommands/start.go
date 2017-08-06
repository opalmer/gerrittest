package cmd

import (
	"encoding/json"
	"errors"
	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString("json")
		if err != nil {
			return err
		}
		if path == "" {
			return errors.New("--json not provided")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}

		image, err := cmd.Flags().GetString("image")
		if err != nil {
			return err
		}

		portHTTP, err := cmd.Flags().GetUint16("port-http")
		if err != nil {
			return err
		}

		portSSH, err := cmd.Flags().GetUint16("port-ssh")
		if err != nil {
			return err
		}

		docker, err := dockertest.NewClient()
		if err != nil {
			return err
		}

		cfg := gerrittest.NewConfig()
		cfg.Image = image
		cfg.PortHTTP = portHTTP
		cfg.PortSSH = portSSH
		svc := gerrittest.NewService(docker, cfg)
		admin, helpers, err := svc.Run()
		if err != nil {
			return err
		}
		client, err := helpers.GetSSHClient(admin)
		if err != nil {
			return err
		}
		defer client.Close()
		version, err := client.Version()
		if err != nil {
			return err
		}

		spec := &gerrittest.ServiceSpec{
			Admin:     admin,
			Container: svc.Service.Container.ID(),
			Version:   version,
			SSH:       helpers.SSH,
			HTTP:      helpers.HTTP,
		}
		data, err := json.Marshal(spec)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(path, data, 0600)
	},
}

func init() {
	Stop.Flags().String(
		"json", "",
		"The json file containing the container to stop.")
}
