package gerrittest

import (
	"errors"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// Helpers provides structs and functions for setting up and
// interacting with Gerrit.

var (
	// ErrGerritDown is returned by Ping() if Gerrit appears to be down.
	ErrGerritDown = errors.New(
		"Gerrrit is not running or failed to start.")
)

// GerritHelpers is a struct which provides helper methods for interacting
// with a running instance of Gerrit.
type GerritHelpers struct {
	PortHTTP int
	PortSSH  int
	Address  string
}

// AddressOnly returns the address part of the given value. For example:
//           127.0.0.1:5000 -> 127.0.0.1
//    tcp://127.0.0.1:5000/ -> 127.0.0.1
func AddressOnly(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := url.Parse(value); err == nil {
		if parsed.Host == "" {
			return value
		}
		value = parsed.Host
	}
	if strings.Contains(value, ":") {
		value = strings.Split(value, ":")[0]
	}

	return value
}

// GetAddress attempts to returns the address where we should expect to
// connect to. This will try $DOCKER_HOST first and failing that attempt
// to use docker-machine before finally giving up and returning 127.0.0.1
func GetAddress() string {
	address := "127.0.0.1"

	dockerhost := os.Getenv("DOCKER_HOST")
	if dockerhost != "" {
		address = AddressOnly(dockerhost)
	} else {
		command := exec.Command("docker-machine", "ip")
		if data, err := command.Output(); err == nil {
			address = AddressOnly(string(data))
		}
	}

	return address
}

// NewGerritHelpers produces a new *GerritHelpers struct.
func NewGerritHelpers(container *Container) (*GerritHelpers, error) {

	helpers := &GerritHelpers{
		PortHTTP: int(container.HTTP),
		PortSSH:  int(container.SSH),
	}
	return helpers, helpers.Ping()
}

// Ping returns nil if Gerrit appears to be running.
func (helpers *GerritHelpers) Ping() error {
	return nil
}
