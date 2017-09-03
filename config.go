package gerrittest

import (
	"context"
	"os"

	"github.com/opalmer/dockertest"
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

	// RepoRoot is the root of the git repository. If this field
	// is blank then a temporary path will be used by the Gerrit
	// struct.
	RepoRoot string

	// Context is used internally when starting or managing
	// containers and processes. If no context is provided then
	// context.Background() will be used.
	Context context.Context

	// PrivateKey is the path to the private key to insert into
	// Gerrit. If a path is not provided then a private key will
	// be generated automatically.
	PrivateKey string

	// Username is the name of the Gerrit admin account to create. By default
	// this will be 'admin' unless otherwise specified.
	Username string

	// Password is the password to create for the Gerrit admin user. If not
	// provided one will be randomly generated for you after the container
	// starts.
	Password string
}

// NewConfig produces a *Config struct with reasonable defaults.
func NewConfig() *Config {
	image := DefaultImage
	if value, set := os.LookupEnv(DefaultImageEnvironmentVar); set {
		image = value
	}
	return &Config{
		Image:    image,
		PortSSH:  dockertest.RandomPort,
		PortHTTP: dockertest.RandomPort,
		RepoRoot: "",
		Username: "admin",
		Password: "",
	}
}
