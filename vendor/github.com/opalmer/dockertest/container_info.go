package dockertest

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

const timeNotSet = "0001-01-01T00:00:00Z"

var (
	// ErrPortNotFound is returned by ContainerInfo.Port if we're unable
	// to find a matching port on the c.
	ErrPortNotFound = errors.New("the requested port could not be found")

	// ErrContainerNotRunning is returned by Started() if the Container
	// was never started.
	ErrContainerNotRunning = errors.New("container not running")

	// ErrContainerStillRunning is returned by Finished() if the Container
	// is still running.
	ErrContainerStillRunning = errors.New("container still running")
)

// ContainerInfo provides a wrapper around information
type ContainerInfo struct {
	JSON     types.ContainerJSON
	Data     types.Container
	State    *types.ContainerState
	Warnings []string
	client   *DockerClient
}

func (c *ContainerInfo) String() string {
	return fmt.Sprintf("{image:%s, id:%s, status:%s}", c.Data.Image, c.Data.ID, c.Data.Status)
}

// GetLabel will return the value of the given label or "" if it does
// not exist. The boolean indicates if the label exists at all.
func (c *ContainerInfo) GetLabel(name string) (string, bool) {
	value, set := c.Data.Labels[name]
	return value, set
}

// HasLabel returns true if the provided label exists and is equal
// to the provided value.
func (c *ContainerInfo) HasLabel(name string, value string) bool {
	current, set := c.GetLabel(name)
	return set && value == current
}

// Refresh will refresh the data present on this struct.
func (c *ContainerInfo) Refresh() error {
	updated, err := c.client.ContainerInfo(context.Background(), c.ID())
	if err != nil {
		return err
	}
	*c = *updated
	return nil
}

// address makes its best effort to determine the ip for the given address,
// port and protocol.
func (c *ContainerInfo) address(ip string, port uint16, protocol Protocol) (string, error) {
	if ip != "0.0.0.0" {
		return ip, nil
	}
	if env, set := os.LookupEnv("DOCKER_URL"); set {
		// If there's a url defined then we'll use that
		// since it's the most likely to work in various
		// scenarios (docker-machine, local docker and remote docker service)
		parsed, err := url.Parse(env)
		if err != nil {
			return "", err
		}
		host := strings.Split(parsed.Host, ":")
		if host[0] != "" {
			return host[0], nil
		}
	}

	// In the majority of cases 127.0.0.1 will be a safe bet if
	// DOCKER_URL is not set. We could try connecting to the port
	// but we don't know if the socket is listening yet and we also
	// can't be 100% certain 127.0.0.1 is connect.
	return "127.0.0.1", nil
}

// Port will return types.Port for the requested internal port. Note, attempts
// will be made to correct the address before returning. If $DOCKER_URL is not
// set however 127.0.0.1 will be returned if a specific IP was not provided by
// Docker.
func (c *ContainerInfo) Port(internal int) (*Port, error) {
	for _, port := range c.Data.Ports {
		if port.PrivatePort == uint16(internal) {
			protocol := ProtocolTCP
			if port.Type == "udp" {
				protocol = ProtocolUDP
			}
			address, err := c.address(port.IP, port.PublicPort, protocol)
			if err != nil {
				return nil, err
			}
			return &Port{
				Private:  port.PrivatePort,
				Public:   port.PublicPort,
				Protocol: protocol,
				Address:  address,
			}, nil
		}
	}
	return nil, ErrPortNotFound
}

// ID is a shortcut function to return the Container's id
func (c *ContainerInfo) ID() string {
	return c.Data.ID
}

// Started returns the time the Container was started at.
func (c *ContainerInfo) Started() (time.Time, error) {
	if c.State.StartedAt == timeNotSet {
		return time.Unix(0, 0).UTC(), ErrContainerNotRunning
	}
	return time.Parse(time.RFC3339Nano, c.State.StartedAt)
}

// Finished returns the time the Container finished running.
func (c *ContainerInfo) Finished() (time.Time, error) {
	if c.State.FinishedAt == timeNotSet {
		return time.Unix(0, 0).UTC(), ErrContainerStillRunning
	}
	return time.Parse(time.RFC3339Nano, c.State.FinishedAt)
}

// Elapsed returns how long the Container has been running or had run if
// the Container has stopped.
func (c *ContainerInfo) Elapsed() (time.Duration, error) {
	started, err := c.Started()
	if err != nil {
		if err == ErrContainerNotRunning {
			return time.Second * 0, nil
		}
		return time.Second * 0, err
	}

	finished, err := c.Finished()
	if err != nil {
		if err == ErrContainerStillRunning {
			return time.Since(started), nil
		}
		return time.Second * 0, nil
	}
	return finished.Sub(started), nil
}
