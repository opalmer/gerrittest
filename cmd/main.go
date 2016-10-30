package main

import (
	"os"
	//"context"

	//"github.com/docker/docker/client"
	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
)

var (
	// Command represents the base command when called without
	// any subcommands
	Command = &cobra.Command{
		Use: "gerrittest",
		Short: "A command line tool for running Gerrit in docker " +
			"for testing."}

	// ShowCommand shows information about running containers
	ShowCommand = &cobra.Command{
		Use:   "show",
		Short: "Shows information about running containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := setup(cmd)
			if err != nil {
				return err
			}

			containers, err := client.Containers()
			if err != nil {
				return err
			}

			for _, container := range containers {

				entry := log.WithFields(log.Fields{
					"id":   container.ID,
					"http": container.HTTP,
					"ssh":  container.SSH,
				})

				entry.Info()
			}

			return nil
		}}
)

func setup(cmd *cobra.Command) (*gerrittest.DockerClient, error) {
	if cmd.Flag("log-level").Changed {
		resolved, err := log.ParseLevel(cmd.Flag("log-level").Value.String())
		if err != nil {
			return nil, err
		}
		log.SetLevel(resolved)
	}

	client, err := gerrittest.NewDockerClient()
	return client, err
}

func init() {
	persistent := Command.PersistentFlags()
	persistent.String(
		"image", "opalmer/gerrittest:latest",
		"The name of the image that should be run.")
	persistent.String(
		"log-level", "", "Override the default log level.")
	Command.AddCommand(ShowCommand)
}

func main() {
	if err := Command.Execute(); err != nil {
		log.WithError(err).Error()
		os.Exit(1)
	}
}
