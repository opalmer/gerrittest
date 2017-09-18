package dockertest

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// ClientInput is used to provide inputs to the RunContainer function.
type ClientInput struct {
	Image       string
	Ports       *Ports
	Labels      map[string]string
	Environment []string

	// Fields provided for the purposes of filtering containers.
	Since     string
	Before    string
	Status    string
	All       bool
	OlderThan time.Duration
}

// SetLabel will add set the provided label key to the provided value.
func (i *ClientInput) SetLabel(key string, value string) {
	i.Labels[key] = value
}

// RemoveLabel will remove the specified label
func (i *ClientInput) RemoveLabel(key string) {
	delete(i.Labels, key)
}

// ContainerConfig will return a *c.Config struct which may
// be passed to the ContainerCreate() API call.
func (i *ClientInput) ContainerConfig() *container.Config {
	return &container.Config{
		Image:  i.Image,
		Labels: i.Labels,
		Env:    i.Environment,
	}
}

// AddEnvironmentVar adds an environment variable
func (i *ClientInput) AddEnvironmentVar(key string, value string) {
	i.Environment = append(
		i.Environment, fmt.Sprintf("%s=%s", key, value))
}

// FilterArgs converts *ClientInput into a filters.Args struct
// which may be used with the docker client directly.
func (i *ClientInput) FilterArgs() filters.Args {
	args := filters.NewArgs()

	if i.Image != "" {
		args.Add("ancestor", i.Image)
	}

	for key, value := range i.Labels {
		args.Add("label", fmt.Sprintf("%s=%s", key, value))
	}

	if i.Status != "" {
		args.Add("status", i.Status)
	}

	return args
}

// NewClientInput produces a *ClientInput struct.
func NewClientInput(image string) *ClientInput {
	input := &ClientInput{
		Image:  image,
		Ports:  NewPorts(),
		Labels: map[string]string{},
	}
	input.SetLabel("dockertest", "1")
	return input
}
