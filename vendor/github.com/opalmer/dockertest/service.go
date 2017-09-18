package dockertest

import (
	"context"
	"errors"

	"github.com/crewjam/errset"
)

var (
	// ErrInputNotProvided is returned by Service.Run if the Input field
	// is not provided.
	ErrInputNotProvided = errors.New("input field not provided")

	// ErrContainerNotStarted is returned by Terminate() if the container
	// was never started.
	ErrContainerNotStarted = errors.New("container not started")
)

// PingInput is used to provide inputs to a Ping function.
type PingInput struct {
	Service   *Service
	Container *ContainerInfo
}

// Ping is a function that's used to ping a service before returning from
// Service.Run. Any errors produced by ping will cause the associated
// Container to be removed.
type Ping func(*PingInput) error

// Service is a struct used to run and manage a Container for a specific
// service.
type Service struct {
	// Name is an optional name that may be used for tracking a service. This
	// field is not used by dockertest.
	Name string

	// Ping is a function that may be used to wait for the service
	// to come up before returning. If this function is specified
	// and it return an error Terminate() will be automatically
	// called. This function is called by Run() before returning.
	Ping Ping

	// Input is used to control the inputs to Run()
	Input *ClientInput

	// Client is the docker client.
	Client *DockerClient

	// Container will container information about the running container
	// once Run() has finished.
	Container *ContainerInfo
}

// Run will run the Container.
func (s *Service) Run() error {
	if s.Input == nil {
		return ErrInputNotProvided
	}

	info, err := s.Client.RunContainer(context.Background(), s.Input)
	if err != nil {
		return err
	}
	s.Container = info

	if s.Ping != nil {
		input := &PingInput{
			Service:   s,
			Container: info,
		}
		if err := s.Ping(input); err != nil {
			errs := errset.ErrSet{}
			errs = append(errs, err)
			errs = append(errs, s.Terminate())
			return errs.ReturnValue()
		}
	}

	return nil
}

// Terminate terminates the Container and returns.
func (s *Service) Terminate() error {
	if s.Container == nil {
		return ErrContainerNotStarted
	}
	return s.Client.RemoveContainer(context.Background(), s.Container.ID())
}
