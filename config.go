package gerrittest

import (
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

// Config is used to tell the *Service struct what setup steps
// to perform, where to listen for services, etc.
type Config struct {
	// Image is the name of docker image to run.
	Image string

	// PortSSH is the port to expose the SSH service on.
	PortSSH uint16

	// PortHTTP is the port to expose the HTTP service on.
	PortHTTP uint16

	// CreateAdmin when true will cause a default admin
	// account to be created.
	CreateAdmin bool

	// Keep indicates if the container should be kept around
	// after we're done and/or after failure.
	Keep bool

	// PublicKey is the key to use to access the service as
	// admin.
	PublicKey ssh.PublicKey

	PrivateKeyPath string
}

// NewConfig produces a *Config struct with reasonable
// default values.
func NewConfig() *Config {
	return &Config{
		Image:          "opalmer/gerrittest:latest",
		PortSSH:        dockertest.RandomPort,
		PortHTTP:       dockertest.RandomPort,
		CreateAdmin:    true,
		Keep:           false,
		PublicKey:      nil,
		PrivateKeyPath: "",
	}
}
