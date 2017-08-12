package cmd

import (
	"encoding/json"
	"errors"
	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKeyPath, err := cmd.Flags().GetString("private-key")
		if err != nil {
			return err
		}

		var publicKey ssh.PublicKey
		if privateKeyPath != "" {
			public, err := gerrittest.ReadSSHPrivateKey(privateKeyPath)
			publicKey = public
			if err != nil {
				return err
			}

		} else {
			public, private, err := gerrittest.GenerateSSHKey()
			publicKey = public
			if err != nil {
				return err
			}
			privateKeyFile, err := ioutil.TempFile("", "")
			if err != nil {
				return err
			}
			privateKeyPath = privateKeyFile.Name()

			if err := gerrittest.WritePrivateKey(private, privateKeyFile.Name()); err != nil {
				return err
			}
		}

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
		cfg.PublicKey = publicKey
		cfg.PrivateKeyPath = privateKeyPath
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
	Start.Flags().StringP(
		"private-key", "i", "",
		"If provided then use this private key instead of generating one.")
}
