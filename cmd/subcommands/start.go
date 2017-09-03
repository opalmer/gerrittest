package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
)

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup timeout and Ctrl+C handling.
		timeout := getDuration(cmd, "timeout")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		go func() {
			defer cancel()
			for range interrupts {
				return
			}
		}()

		config := gerrittest.NewConfig()
		config.Image = getString(cmd, "image")
		config.PortSSH = getUInt16(cmd, "port-http")
		config.PortHTTP = getUInt16(cmd, "port-http")
		config.PrivateKey = getString(cmd, "private-key")
		config.Password = getString(cmd, "password")
		config.Context = ctx
		config.SkipSetup = getBool(cmd, "start-only")
		config.SkipCleanup = getBool(cmd, "no-cleanup")

		gerrit, err := gerrittest.New(config)
		if err != nil {
			return err
		}

		return jsonOutput(cmd, gerrit)
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
		"image", gerrittest.DefaultImage,
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
	Start.Flags().StringP(
		"password", "p", "",
		"If provided then use this value for the admin password instead "+
			"of generating one.")
}
