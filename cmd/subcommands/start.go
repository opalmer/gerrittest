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

func addStartFlags(cmd *cobra.Command) {
	cmd.Flags().Duration(
		"timeout", time.Minute*2,
		"The maximum amount of time to wait for the service to come up.")
	cmd.Flags().BoolP(
		"no-cleanup", "n", false,
		"If provided then do not cleanup the container on failure. "+
			"Useful when debugging changes to the docker image.")
	cmd.Flags().String(
		"json", "",
		"The location to write information about the service to. Any "+
			"existing content will be overwritten.")
	cmd.Flags().String(
		"image", gerrittest.DefaultImage,
		"The Docker image to spin up Gerrit.")
	cmd.Flags().Uint16(
		"port-http", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
	cmd.Flags().Uint16(
		"port-ssh", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
	cmd.Flags().StringP(
		"private-key", "i", "",
		"If provided then use this private key instead of generating one.")
	cmd.Flags().Bool(
		"start-only", false,
		"If provided just start the container, don't setup anything else.")
	cmd.Flags().StringP(
		"password", "p", "",
		"If provided then use this value for the admin password instead "+
			"of generating one.")
	cmd.Flags().String(
		"project", gerrittest.ProjectName,
		"The name of the project to create in Gerrit. This will "+
			"also be used for the remote repo name.")
}

func newStartConfig(cmd *cobra.Command) *gerrittest.Config {
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
	config.PortSSH = getUInt16(cmd, "port-ssh")
	config.PortHTTP = getUInt16(cmd, "port-http")
	config.PrivateKeyPath = getString(cmd, "private-key")
	config.Password = getString(cmd, "password")
	config.Context = ctx
	config.SkipSetup = getBool(cmd, "start-only")
	if getBool(cmd, "no-cleanup") {
		config.CleanupPrivateKey = false
		config.CleanupContainer = false
	}

	return config
}

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := newStartConfig(cmd)
		gerrit, err := gerrittest.New(cfg)
		if err != nil {
			return gerrit.Destroy()
		}
		return jsonOutput(cmd, gerrit)
	},
}

func init() {
	addStartFlags(Start)
}
