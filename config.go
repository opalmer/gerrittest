package gerrittest

import (
	"context"
	"os"
	"time"

	"github.com/go-ini/ini"
	"github.com/opalmer/dockertest"
	log "github.com/sirupsen/logrus"
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

	// Timeout is used to timeout commands and other contextual operations.
	Timeout time.Duration `json:"timeout"`

	// GitConfig contains key/value pairs to pass to 'git config'
	GitConfig map[string]string `json:"git"`

	// Context is used internally when starting or managing
	// containers and processes. If no context is provided then
	// context.Background() will be used.
	Context context.Context `json:"-"`

	// SSHKeys store information about one or more ssh keys.
	SSHKeys []*SSHKey `json:"ssh_keys"`

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

	// CleanupContainer when true will cause the cleanup steps to destroy
	// the container running Gerrit. This defaults to true.
	CleanupContainer bool `json:"cleanup_container"`
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
		Timeout:  time.Minute * 5,
		GitConfig: map[string]string{
			"user.name":  "admin",
			"user.email": "admin@localhost",
		},
		Context:          context.Background(),
		SSHKeys:          []*SSHKey{},
		Username:         "admin",
		Password:         "",
		SkipSetup:        false,
		CleanupContainer: true,
	}
}

const (
	accessHeads   = `access "refs/heads/*"`
	labelVerified = `label "Verified"`
)

// projectConfig is an internal struct which is used to edit the existing
// configuration.
type projectConfig struct {
	log *log.Entry
	ini *ini.File
}

func (c *projectConfig) modifyVerifiedTag() {
	logger := c.log.WithField("phase", "create-verified-tag")
	logger.Debug()

	section := c.ini.Section(labelVerified)
	section.Key("function").SetValue("MaxWithBlock")
	section.Key("defaultValue").SetValue("0")
	value := section.Key("value")
	value.SetValue("-1 Fails")
	value.AddShadow("0 No Score")  // nolint: errcheck
	value.AddShadow("+1 Verified") // nolint: errcheck
}

func (c *projectConfig) modifyAccess() {
	logger := c.log.WithField("phase", "modify-access")
	logger.Debug()

	section := c.ini.Section(accessHeads)
	value := section.Key("label-Verified")
	value.SetValue("-1..+1 group Administrators")
	value.AddShadow("-1..+1 group Project Owners") // nolint: errcheck
}

func (c *projectConfig) write(path string) error {
	c.modifyVerifiedTag()
	c.modifyAccess()
	c.log.WithField("phase", "write").Debug()
	return c.ini.SaveTo(path)
}

// newProjectConfig produces a projectConfig struct.
func newProjectConfig(path string) (*projectConfig, error) {
	logger := log.WithField("cmp", "project-config")
	options := ini.LoadOptions{AllowShadows: true}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.WithField("action", "new").Debug()
		cfg, err := ini.LoadSources(options, []byte(""))
		if err != nil {
			return nil, err
		}
		return &projectConfig{log: logger, ini: cfg}, nil
	}

	logger = logger.WithField("path", path)
	logger.WithField("action", "load").Debug()
	cfg, err := ini.LoadSources(options, path)
	if err != nil {
		return nil, err
	}
	return &projectConfig{log: logger, ini: cfg}, nil
}
