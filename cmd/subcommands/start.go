package cmd

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"time"
	"encoding/json"
)

func getSSHKeys(cmd *cobra.Command) (ssh.PublicKey, *rsa.PrivateKey, string, error) {
	privateKeyPath, err := cmd.Flags().GetString("private-key")
	if err != nil {
		return nil, nil, "", err
	}

	if privateKeyPath == "" {
		public, private, err := gerrittest.GenerateSSHKeys()
		if err != nil {
			return nil, nil, "", err
		}
		return public, private, "", err
	}
	public, private, err := gerrittest.ReadSSHKeys(privateKeyPath)

	return public, private, privateKeyPath, nil
}

// NewConfigFromCommand converts a command to a config struct.
func NewConfigFromCommand(cmd *cobra.Command) (*gerrittest.Config, error) {
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		return nil, err
	}

	portHTTP, err := cmd.Flags().GetUint16("port-http")
	if err != nil {
		return nil, err
	}

	portSSH, err := cmd.Flags().GetUint16("port-ssh")
	if err != nil {
		return nil, err
	}

	noCleanup, err := cmd.Flags().GetBool("no-cleanup")
	if err != nil {
		return nil, nil
	}

	return &gerrittest.Config{
		Image:            image,
		PortSSH:          portSSH,
		PortHTTP:         portHTTP,
		CleanupOnFailure: noCleanup == false,
	}, nil
}

func jsonOutput(cmd *cobra.Command, spec *gerrittest.ServiceSpec) error {
	data, err := json.MarshalIndent(spec, "", " ")
	if err != nil {
		return err
	}

	jsonPath, err := cmd.Flags().GetString("json")
	if err != nil {
		return err
	}
	if jsonPath != "" {
		if err := os.MkdirAll(filepath.Dir(jsonPath), 0700); err != nil {
			return err
		}
		return ioutil.WriteFile(jsonPath, data, 0600)
	}
	fmt.Println(string(data))
	return nil
}

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := NewConfigFromCommand(cmd)
		if err != nil {
			return err
		}
		// Setup timeout and Ctrl+C handling.
		timeout, err := cmd.Flags().GetDuration("timeout")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		go func() {
			for range interrupts {
				cancel()
			}
		}()

		spec := &gerrittest.ServiceSpec{Admin: &gerrittest.User{}}
		service, err := gerrittest.Start(ctx, cfg)
		if err != nil {
			return err
		}

		spec.HTTP = service.HTTPPort
		spec.SSH = service.SSHPort
		spec.Container = service.Container.ID()

		startonly, err := cmd.Flags().GetBool("start-only")
		if startonly {
			return nil
		}

		client, err := service.HTTPClient()
		if err != nil {
			return err
		}

		// Hitting /login/ will produce a cookie that can be used
		// for authenticated requests. Also, this first request
		// causes the first account to be created which happens
		// to be the admin account.
		if err := client.Login(); err != nil {
			return err
		}

		account, err := client.GetAccount("self")
		if err != nil {
			return err
		}
		spec.Admin.Login = account.Username

		password, err := client.GeneratePassword()
		if err != nil {
			return err
		}
		spec.Admin.Password = password


		return jsonOutput(cmd, spec)
	},
}

func init() {
	Start.Flags().Duration(
		"timeout", time.Minute*2,
		"The maximum amount of time to wait for the service to come up.")
	Start.Flags().BoolP(
		"no-cleanup", "n", false,
		"If provided then do not cleanup the container on failure. "+
			"Useful when debugging changes to the docker image.")
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
	Start.Flags().Bool(
		"start-only", false,
		"If provided just start the container, don't setup anything else.")
}
