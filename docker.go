package gerrittest

import (
	"context"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var (
	// ErrPublicPortsMissing is returned when a container does not appear
	// to expose the expected internal ports.
	ErrPublicPortsMissing = errors.New(
		"Failed to determine public ports of container.")
)

// Container wraps the standard container type.
type Container struct {
	types.Container
	HTTP uint16 `json:"port_http"`
	SSH  uint16 `json:"port_ssh"`
}

// DockerClient provides a wrapper for the standard docker client
type DockerClient struct {
	Docker *client.Client
	log    *log.Entry
}

// NewDockerClient returns a *DockerClient struct
func NewDockerClient() (*DockerClient, error) {
	logEntry := log.WithField("phase", "new-client")
	logEntry.Debug("Constructing new client")
	docker, err := client.NewEnvClient()
	if err != nil {
		logEntry.WithError(err).Error()
		return nil, err
	}

	cli := &DockerClient{
		Docker: docker,
		log:    log.WithField("phase", "docker")}

	return cli, nil
}

// Containers returns a list of gerrittest containers.
func (client *DockerClient) Containers() ([]*Container, error) {
	args := filters.NewArgs()
	args.Add("ancestor", "opalmer/gerrittest")

	options := types.ContainerListOptions{Filter: args}
	containers, err := client.Docker.ContainerList(context.Background(), options)

	if err != nil {
		client.log.WithError(err).Error()
		return []*Container{}, err
	}
	output := []*Container{}

	for _, container := range containers {
		if wrapped, err := NewContainer(container); err == nil {
			output = append(output, wrapped)
			continue
		}
		client.log.WithError(err).WithField("id", container.ID).Warn()
	}

	return output, err
}

// NewContainer returns a *Container struct for the given container.
func NewContainer(container types.Container) (*Container, error) {
	http := uint16(0)
	ssh := uint16(0)
	for _, port := range container.Ports {
		switch port.PrivatePort {
		case InternalSSHPort:
			ssh = port.PublicPort
		case InternalHTTPPort:
			http = port.PublicPort
		}
	}
	if http != 0 && ssh != 0 {
		return &Container{container, http, ssh}, nil
	}
	return nil, ErrPublicPortsMissing
}
