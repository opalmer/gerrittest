package gerrittest

import (
	"context"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

// Gerrit is the central struct which combines multiple components
// of the gerrittest projects. Use New() to construct this struct.
type Gerrit struct {
	log            *log.Entry
	Config         *Config
	Service        *Service
	HTTP           *HTTPClient
	HTTPPort       *dockertest.Port
	SSH            *SSHClient
	SSHPort        *dockertest.Port
	PrivateKey     ssh.Signer
	PublicKey      ssh.PublicKey
	PrivateKeyPath string
	Username       string
	Password       string
}

// startContainer starts the docker container containing Gerrit.
func (g *Gerrit) startContainer() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "start-container",
	})
	logger.Debug()
	service, err := Start(g.Config.Context, g.Config)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}
	g.Service = service
	g.SSHPort = service.SSHPort
	g.HTTPPort = service.HTTPPort
	return nil
}

// setupSSHKey loads or generates an SSH key.
func (g *Gerrit) setupSSHKey() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "ssh",
	})
	logger.Debug()

	if g.Config.PrivateKey != "" {
		entry := logger.WithFields(log.Fields{
			"action": "read",
			"path":   g.Config.PrivateKey,
		})
		entry.Debug()
		public, private, err := ReadSSHKeys(g.Config.PrivateKey)
		if err != nil {
			entry.WithError(err).Error()
			return err
		}
		g.PrivateKey = private
		g.PublicKey = public
		return nil
	}
	entry := logger.WithFields(log.Fields{
		"action": "generate",
	})

	private, err := GenerateRSAKey()
	if err != nil {
		entry.WithError(err).Error()
		return err
	}

	file, err := ioutil.TempFile("", "gerrittest-id_rsa-")
	entry = entry.WithField("path", file.Name())
	if err != nil {
		entry.WithError(err).Error()
		return err
	}

	defer file.Close()
	if err := WriteRSAKey(private, file); err != nil {
		entry.WithError(err).Error()
		return err
	}

	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		entry.WithError(err).Error()
		return err
	}
	g.PrivateKey = signer
	g.PublicKey = signer.PublicKey()
	g.PrivateKeyPath = file.Name()
	return nil
}

func (g *Gerrit) setupHTTPClient() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "http-client",
	})

	client := NewHTTPClient(g.Service, g.Username)

	logger.WithField("action", "login").Debug()
	if err := client.Login(); err != nil {
		logger.WithError(err).Error()
		return err
	}

	logger.WithField("action", "insert-key").Debug()
	if err := client.InsertPublicKey(g.PublicKey); err != nil {
		logger.WithError(err).Error()
		return err
	}

	if g.Password == "" {
		logger.WithField("action", "generate-password").Debug()
		generated, err := client.GeneratePassword()
		if err != nil {
			logger.WithError(err).Error()
			return err
		}

		g.Password = generated
		return nil
	}

	g.HTTP = client
	logger.WithField("action", "set-password").Debug()
	return client.SetPassword(g.Password)
}

func (g *Gerrit) setupSSHClient() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "ssh-client",
	})
	logger.Debug()

	client, err := NewSSHClient(g.Username, g.PrivateKeyPath, g.SSHPort)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	g.SSH = client
	return nil
}

// Destroy will destroy the container and all associated resources.
func (g *Gerrit) Destroy() error {
	return nil
}

// New constructs and returns a *Gerrit struct after all setup steps have
// been completed. Once this function returns Gerrit will be running in
// a container, an admin user will be created and a git repository will
// be setup pointing at the service in the container.
func New(cfg *Config) (*Gerrit, error) {
	// Use a temp. directory if no repository root was provided.
	if cfg.RepoRoot == "" {
		path, err := ioutil.TempDir("", "gerrittest-")
		if err != nil {
			return nil, err
		}
		cfg.RepoRoot = path
	}

	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	gerrit := &Gerrit{
		log:    log.WithField("cmp", "core"),
		Config: cfg,
	}
	if err := gerrit.setupSSHKey(); err != nil {
		return nil, err
	}
	if err := gerrit.startContainer(); err != nil {
		return nil, err
	}
	if err := gerrit.setupHTTPClient(); err != nil {
		return nil, err
	}
	if err := gerrit.setupSSHClient(); err != nil {
		return nil, err
	}

	return gerrit, nil
}
