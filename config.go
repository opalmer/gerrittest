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
	Image string `json:"image"`

	// PortSSH is the port to expose the SSH service on.
	PortSSH uint16 `json:"port_ssh"`

	// PortHTTP is the port to expose the HTTP service on.
	PortHTTP uint16 `json:"port_http"`

	// RepoRoot is the root of the git repository. If this field
	// is blank then a temporary path will be used by the Gerrit
	// struct.
	RepoRoot string `json:"repo_root"`

	// Context is used internally when starting or managing
	// containers and processes. If no context is provided then
	// context.Background() will be used.
	Context context.Context `json:"-"`

	// PrivateKey is the path to the private key to insert into
	// Gerrit. If a path is not provided then a private key will
	// be generated automatically.
	PrivateKey string `json:"private_key"`

	// Username is the name of the Gerrit admin account to create. By default
	// this will be 'admin' unless otherwise specified.
	Username string `json:"username"`

	// Password is the password to create for the Gerrit admin user. If not
	// provided one will be randomly generated for you after the container
	// starts.
	Password string `json:"password"`

	// SkipSetup when true will cause the container to be started but
	// none of the final setup steps will be performed.
	SkipSetup bool `json:"skip_setup"`

	// SkipCleanup when true allows all cleanup steps to be skipped.
	SkipCleanup bool `json:"skip_cleanup"`
}

// NewConfig produces a *Config struct with reasonable defaults.
func NewConfig() *Config {
	image := DefaultImage
	if value, set := os.LookupEnv(DefaultImageEnvironmentVar); set {
		image = value
	}
	return &Config{
		Image:       image,
		PortSSH:     dockertest.RandomPort,
		PortHTTP:    dockertest.RandomPort,
		RepoRoot:    "",
		Username:    "admin",
		Password:    "",
		SkipSetup:   false,
		SkipCleanup: false,
	}
}
