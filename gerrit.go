package gerrittest

import (
	"context"
	"io/ioutil"
	"os"

	"encoding/json"

	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/crewjam/errset"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

// Gerrit is the central struct which combines multiple components
// of the gerrittest project. Use New() to construct this struct.
type Gerrit struct {
	log             *log.Entry
	cleanRepo       bool
	cleanPrivateKey bool
	Config          *Config          `json:"config"`
	Container       *Container       `json:"container"`
	HTTP            *HTTPClient      `json:"-"`
	HTTPPort        *dockertest.Port `json:"http"`
	SSH             *SSHClient       `json:"-"`
	SSHPort         *dockertest.Port `json:"ssh"`
	Repo            *Repository      `json:"repo"`
	PrivateKey      ssh.Signer       `json:"-"`
	PublicKey       ssh.PublicKey    `json:"-"`
	PrivateKeyPath  string           `json:"private_key_path"`
	Username        string           `json:"username"`
	Password        string           `json:"password"`
}

// startContainer starts the docker container containing Gerrit.
func (g *Gerrit) startContainer() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "start-container",
	})
	logger.Debug()
	container, err := NewContainer(
		g.Config.Context, g.Config.PortHTTP, g.Config.PortSSH, g.Config.Image)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	// Cookies are set based on hostname so we need to be
	// consistent and use 'localhost' if we're working with
	// 127.0.0.1.
	if container.HTTP.Address == "127.0.0.1" {
		container.HTTP.Address = "localhost"
	}

	g.Container = container
	g.SSHPort = container.SSH
	g.HTTPPort = container.HTTP

	return nil
}

// setupSSHKey loads or generates an SSH key.
func (g *Gerrit) setupSSHKey() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "ssh-key",
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

	defer file.Close() // nolint: errcheck
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

	client, err := NewHTTPClient(g.HTTPPort.Address, g.HTTPPort.Public, g.Username)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

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

func (g *Gerrit) setupRepo() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "repo",
	})
	logger.Debug()

	path := g.Config.RepoRoot
	if path == "" {
		tmppath, err := ioutil.TempDir("", "gerrittest-")
		if err != nil {
			return err
		}
		path = tmppath
	}
	cfg, err := newRepositoryConfig(path, g.PrivateKeyPath)
	if err != nil {
		return err
	}
	repo, err := NewRepository(cfg)
	g.Repo = repo
	return err
}

// WriteJSONFile takes the current struct and writes the data to disk
// as json.
func (g *Gerrit) WriteJSONFile(path string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

// Destroy will destroy the container and all associated resources. Custom
// private keys or repositories will not be cleaned up.
func (g *Gerrit) Destroy() error {
	if g.Config.SkipCleanup {
		return nil
	}

	g.log.WithField("phase", "destroy").Debug()
	errs := errset.ErrSet{}
	if g.SSH != nil {
		errs = append(errs, g.SSH.Close())
	}
	if g.Container != nil {
		errs = append(errs, g.Container.Terminate())
	}
	if g.cleanRepo && g.Repo != nil {
		errs = append(errs, g.Repo.Remove())
	}
	if g.cleanPrivateKey && g.PrivateKeyPath != "" {
		errs = append(errs, os.Remove(g.PrivateKeyPath))
	}
	return errs.ReturnValue()
}

// New constructs and returns a *Gerrit struct after all setup steps have
// been completed. Once this function returns Gerrit will be running in
// a container, an admin user will be created and a git repository will
// be setup pointing at the service in the container.
func New(cfg *Config) (*Gerrit, error) {
	username := cfg.Username
	if username == "" {
		username = "admin"
	}

	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	gerrit := &Gerrit{
		log:             log.WithField("cmp", "core"),
		Config:          cfg,
		cleanRepo:       cfg.RepoRoot == "",
		cleanPrivateKey: cfg.PrivateKey == "",
		Username:        username,
	}
	if err := gerrit.setupSSHKey(); err != nil {
		return gerrit, err
	}
	if err := gerrit.startContainer(); err != nil {
		return gerrit, err
	}

	if cfg.SkipSetup {
		return gerrit, nil
	}

	if err := gerrit.setupHTTPClient(); err != nil {
		return gerrit, err
	}
	if err := gerrit.setupSSHClient(); err != nil {
		return gerrit, err
	}
	if err := gerrit.setupRepo(); err != nil {
		return gerrit, err
	}

	return gerrit, nil
}

// NewFromJSON reads information from a json file and returns a *Gerrit
// struct.
func NewFromJSON(path string) (*Gerrit, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	gerrit := &Gerrit{}
	return gerrit, json.Unmarshal(data, gerrit)
}
