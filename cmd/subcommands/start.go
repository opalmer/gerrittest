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
	path, err := cmd.Flags().GetString("json")
	if err != nil {
		return nil, err
	}
	if path == "" {
		file, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, err
		}
		path = file.Name()
		if err := file.Close(); err != nil {
			return nil, nil
		}
		fmt.Println(path)

		if err := os.Remove(path); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

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

		service, err := gerrittest.Start(ctx, cfg)
		if err != nil {
			return err
		}
		client, err := service.HTTPClient()
		if err != nil {
			return err
		}
		return client.Login("admin")
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
}
