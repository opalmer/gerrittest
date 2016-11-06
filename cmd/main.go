package main

import (
	"fmt"
	"os"

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

	// ListCommand shows information about running containers
	ListCommand = &cobra.Command{
		Use:   "list",
		Short: "Lists information about running containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newdockerclient(cmd)
			if err != nil {
				return err
			}

			containers, err := client.Containers()
			if err != nil {
				return err
			}

			for _, container := range containers {
				printinfo(container)
			}

			return nil
		}}

	// RunCommand is the command used to run a container
	RunCommand = &cobra.Command{
		Use:   "run",
		Short: "Runs Gerrit in a docker container and returns information about it (id, ssh port, http port)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newdockerclient(cmd)
			if err != nil {
				return err
			}

			image := gerrittest.DefaultImage
			if cmd.Flag("image").Changed {
				image = cmd.Flag("image").Value.String()
			}

			// Create the container
			created, err := client.RunGerrit(&gerrittest.RunGerritInput{
				Image:    image,
				PortHTTP: cmd.Flag("port-http").Value.String(),
				PortSSH:  cmd.Flag("port-ssh").Value.String()})
			if err != nil {
				return err
			}
			printinfo(created)
			return nil
		}}

	// DestroyCommand is used to destroy containers related to gerrrittest.
	DestroyCommand = &cobra.Command{
		Use:   "destroy",
		Short: "Destroys containers running Gerrit. By default all containers will be destroyed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newdockerclient(cmd)
			if err != nil {
				return err
			}
			ids := []string{}
			for _, value := range args {
				container, err := client.GetContainer(value)
				if err != nil {
					log.WithError(err).Errorf(
						"No container found for '%s'", value)
					continue
				}
				ids = append(ids, container.ID)
			}

			if len(args) == 0 {
				containers, err := client.Containers()
				if err != nil {
					return err
				}
				for _, container := range containers {
					ids = append(ids, container.ID)
				}
			}

			for _, id := range ids {
				if err := client.RemoveContainer(id); err != nil {
					log.WithError(err).Errorf("Failed to remove %s", err)
					continue
				}
				fmt.Println(id)
			}

			return nil
		}}
)

func printinfo(container *gerrittest.Container) {
	fmt.Printf(
		"%s %d %d\n", container.ID, container.SSH, container.HTTP)
}

func newdockerclient(cmd *cobra.Command) (*gerrittest.DockerClient, error) {
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
		"log-level", "", "Override the default log level.")

	Command.AddCommand(ListCommand)
	Command.AddCommand(RunCommand)
	Command.AddCommand(DestroyCommand)

	// Run flags
	RunCommand.Flags().Int(
		"port-http", 0,
		"If provided run Gerrit's HTTP service on this port. A random "+
			"port will be chosen otherwise.")
	RunCommand.Flags().Int(
		"port-ssh", 0,
		"If provided run Gerrit's SSH service on this port. A random "+
			"port will be chosen otherwise.")
	RunCommand.Flags().String(
		"image", gerrittest.DefaultImage,
		"The name of the image that should be run.")
}

func main() {
	if err := Command.Execute(); err != nil {
		log.WithError(err).Error()
		os.Exit(1)
	}
}
