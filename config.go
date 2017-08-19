package gerrittest

import (
	"os"

	"github.com/opalmer/dockertest"
)

var (
	// DefaultImage defines the default docker image to use in
	// NewConfig(). This may be overridden with the $GERRITTEST_DOCKER_IMAGE
	// environment variable.
	DefaultImage = "opalmer/gerrittest:2.14.2"
)

// Config is used to tell the *runner struct what setup steps
// to perform, where to listen for services, etc.
type Config struct {
	// Image is the name of docker image to run.
	Image string

	// PortSSH is the port to expose the SSH service on.
	PortSSH uint16

	// PortHTTP is the port to expose the HTTP service on.
	PortHTTP uint16

	// CleanupOnFailure indicates if the container should be kept around
	// after we're done and/or after failure.
	CleanupOnFailure bool
}

// NewConfig produces a *Config struct with reasonable defaults.
func NewConfig() *Config {
	image := DefaultImage
	if value, set := os.LookupEnv("GERRITTEST_DOCKER_IMAGE"); set {
		image = value
	}
	return &Config{
		Image:            image,
		PortSSH:          dockertest.RandomPort,
		PortHTTP:         dockertest.RandomPort,
		CleanupOnFailure: true,
	}
}
