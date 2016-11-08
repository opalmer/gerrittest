package gerrittest

import (
	"context"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	// ErrPublicPortsMissing is returned when a container does not appear
	// to expose the expected internal ports.
	ErrPublicPortsMissing = errors.New(
		"Failed to determine public ports of container.")

	// ErrContainerNotFound is returned by GetContainer if we were
	// unable to find the requested container.
	ErrContainerNotFound = errors.New(
		"Expected to find exactly one container for the given query.")
)

const (
	// DefaultImage is the docker image that's used in tests
	// and by DockerClient if no image is supplied.
	DefaultImage = "opalmer/gerrittest:latest"

	// InternalHTTPPort is the port Gerrit is listening for http requests
	// on inside the container.
	InternalHTTPPort = uint16(8080)

	// InternalSSHPort is the port Gerrit is listening for ssh connection on
	// inside the container.
	InternalSSHPort = uint16(29418)
)

// Container wraps the standard container type.
type Container struct {
	types.Container
	HTTP uint16 `json:"port_http"`
	SSH  uint16 `json:"port_ssh"`
}

// RunGerritInput is used as an input to DockerClient.RunGerrit.
type RunGerritInput struct {
	// Image is the docker image to use when running Gerrit.
	Image string

	// PortHTTP is the local tcp port to expose Gerrit's HTTP interface
	// and API on. If not provided Docker will assign a random port.
	PortHTTP string

	// PortHTTP is the local tcp port to expose Gerrit's SSH server
	// on. If not provided Docker will assign a random port.
	PortSSH string
}

// DockerClient provides a wrapper for the standard docker client
type DockerClient struct {
	Docker *client.Client
	log    *log.Entry
}

// NewDockerClient returns a *DockerClient struct. If you supply "" to
// default image gerrittest.DefaultImage will be used.
func NewDockerClient() (*DockerClient, error) {
	logEntry := log.WithField("phase", "new-client")
	logEntry.Debug("Constructing new client")

	docker, err := client.NewEnvClient()
	if err != nil {
		logEntry.WithError(err).Error()
		return nil, err
	}

	return &DockerClient{
		Docker: docker,
		log:    log.WithField("phase", "docker")}, nil
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

// Containers returns a list of gerrittest containers.
func (client *DockerClient) Containers() ([]*Container, error) {
	args := filters.NewArgs()
	args.Add("ancestor", "opalmer/gerrittest")

	options := types.ContainerListOptions{Filters: args}
	containers, err := client.Docker.ContainerList(context.Background(), options)

	if err != nil {
		client.log.WithError(err).Error()
		return []*Container{}, err
	}
	output := []*Container{}

	for _, entry := range containers {
		if value, set := entry.Labels["gerrittest"]; !set || value != "1" {
			continue
		}
		if wrapped, err := NewContainer(entry); err == nil {
			output = append(output, wrapped)
			continue
		}

		client.log.WithError(err).WithField("id", entry.ID).Warn()
	}

	return output, err
}

// GetContainer retrieves a single container by id.
func (client *DockerClient) GetContainer(id string) (*Container, error) {
	args := filters.NewArgs()
	args.Add("id", id)
	options := types.ContainerListOptions{Filters: args}
	containers, err := client.Docker.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	if len(containers) != 1 {
		return nil, ErrContainerNotFound
	}

	return NewContainer(containers[0])
}

// RunGerrit will run a container using the provided inputs.
func (client *DockerClient) RunGerrit(input *RunGerritInput) (*Container, error) {
	if input == nil {
		input = &RunGerritInput{}
	}

	if input.Image == "" {
		input.Image = DefaultImage
	}

	hostconfig := &container.HostConfig{PublishAllPorts: true}
	portSpecs := []string{}
	if input.PortHTTP != "" {
		portSpecs = append(
			portSpecs, fmt.Sprintf("%s:%d", input.PortHTTP, InternalHTTPPort))
	}

	if input.PortSSH != "" {
		portSpecs = append(
			portSpecs, fmt.Sprintf("%s:%d", input.PortSSH, InternalSSHPort))
	}

	if len(portSpecs) > 0 {
		_, bindings, err := nat.ParsePortSpecs(portSpecs)
		if err != nil {
			return nil, err
		}
		hostconfig.PortBindings = bindings
	}

	logger := client.log.WithField("phase", "run")

	// Create the container
	created, err := client.Docker.ContainerCreate(
		context.Background(), &container.Config{
			Image:  input.Image,
			Labels: map[string]string{"gerrittest": "1"},
		},
		hostconfig, &network.NetworkingConfig{}, "")
	if err != nil {
		return nil, err
	}
	for _, warning := range created.Warnings {
		logger.WithField("container", created.ID).Warn(warning)
	}

	err = client.Docker.ContainerStart(
		context.Background(), created.ID, types.ContainerStartOptions{})

	// Start the container
	if err != nil {
		return nil, err
	}

	return client.GetContainer(created.ID)
}

// RemoveContainer will remove the specified container immediately. This
// is equivalent to `docker rm -f <container id>`.
func (client *DockerClient) RemoveContainer(id string) error {
	return client.Docker.ContainerRemove(
		context.Background(), id,
		types.ContainerRemoveOptions{Force: true})
}

// Helpers returns a GerritHelpers struct for the given container.
func (client *DockerClient) Helpers(input *Container) *GerritHelpers {
	return NewGerritHelpers(input)
}
