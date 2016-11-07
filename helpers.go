package gerrittest

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// Helpers provides structs and functions for setting up and
// interacting with Gerrit.

// GerritHelpers is a struct which provides helper methods for interacting
// with a running instance of Gerrit.
type GerritHelpers struct {
	// PortHTTP is the port that should be used when attempting to connect
	// to Gerrit.
	PortHTTP int

	// PortSSH is the port that should be used when attempting to connect
	// to Gerrit.
	PortSSH int

	// Address is the IP address or hostname that is running the docker
	// container.
	Address string

	// URL is the root url to make API requests to.  It will be
	// something like http://127.0.0.1:8080
	URL string
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
// to use docker-machine before finally giving up and returning 127.0.0.1.
func GetAddress() string {
	address := "127.0.0.1"

	if dockerhost, set := os.LookupEnv("DOCKER_HOST"); set {
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
	address := GetAddress()
	httpport := int(container.HTTP)
	helpers := &GerritHelpers{
		PortHTTP: httpport,
		PortSSH:  int(container.SSH),
		Address:  address,
		URL:      fmt.Sprintf("http://%s:%d", address, httpport)}
	return helpers, helpers.Ping()
}

// Ping returns nil if Gerrit appears to be running. This will ensures that
// we can make a successful HTTP request and that w can connect to ssh.
func (helpers *GerritHelpers) Ping() error {
	// Attempt to connect to http by making a request.
	if _, err := http.Get(helpers.URL); err != nil {
		return err
	}

	_, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", helpers.Address, helpers.PortSSH),
		&ssh.ClientConfig{})
	if strings.Contains(err.Error(), "unable to authenticate") {
		return nil
	}
	return err
}
