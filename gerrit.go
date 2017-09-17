package gerrittest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/crewjam/errset"
	"github.com/opalmer/dockertest"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// ProjectName is used anywhere we need a default value (temp files, default
// field values, etc.
const ProjectName = "gerrittest"

// Gerrit is the central struct which combines multiple components
// of the gerrittest project. Use New() to construct this struct.
type Gerrit struct {
	log        *log.Entry
	Config     *Config          `json:"config"`
	Container  *Container       `json:"container"`
	HTTP       *HTTPClient      `json:"-"`
	HTTPPort   *dockertest.Port `json:"http"`
	SSH        *SSHClient       `json:"-"`
	SSHPort    *dockertest.Port `json:"ssh"`
	PrivateKey ssh.Signer       `json:"-"`
	PublicKey  ssh.PublicKey    `json:"-"`
}

func (g *Gerrit) errLog(logger *log.Entry, err error) error {
	logger.WithError(err).Error()
	return err
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
	g.Config.CleanupPrivateKey = false

	if g.Config.PrivateKeyPath != "" {
		entry := logger.WithFields(log.Fields{
			"action":   "read",
			"key-path": g.Config.PrivateKeyPath,
		})
		entry.Debug()
		public, private, err := ReadSSHKeys(g.Config.PrivateKeyPath)
		if err != nil {
			entry.WithError(err).Error()
			return err
		}
		g.Config.GitConfig["core.sshCommand"] = fmt.Sprintf(
			"ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", g.Config.PrivateKeyPath)
		g.PrivateKey = private
		g.PublicKey = public
		return nil
	}
	entry := logger.WithField("action", "generate")

	private, err := GenerateRSAKey()
	if err != nil {
		return g.errLog(entry, err)
	}

	file, err := ioutil.TempFile("", fmt.Sprintf("%s-id_rsa-", ProjectName))
	entry = entry.WithField("path", file.Name())
	if err != nil {
		return g.errLog(entry, err)
	}

	defer file.Close() // nolint: errcheck
	if err := WriteRSAKey(private, file); err != nil {
		return g.errLog(entry, err)
	}

	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return g.errLog(entry, err)
	}
	g.PrivateKey = signer
	g.PublicKey = signer.PublicKey()
	g.Config.PrivateKeyPath = file.Name()
	g.Config.CleanupPrivateKey = true
	g.Config.GitConfig["core.sshCommand"] = fmt.Sprintf(
		"ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", g.Config.PrivateKeyPath)
	return nil
}

func (g *Gerrit) setupHTTPClient() error { // nolint: gocyclo
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "http-client",
	})

	client, err := NewHTTPClient(g.Config, g.HTTPPort)
	if err != nil {
		return g.errLog(logger, err)
	}

	g.HTTP = client

	logger.WithField("action", "login").Debug()
	if err := g.HTTP.login(); err != nil {
		return g.errLog(logger, err)
	}

	logger.WithField("action", "insert-key").Debug()
	if err := g.HTTP.insertPublicKey(g.PublicKey); err != nil {
		return g.errLog(logger, err)
	}

	// Generate or set the password.
	if g.Config.Password != "" {
		logger = logger.WithField("action", "set-password")
		logger.Debug()
		if _, err := g.HTTP.Gerrit(); err != nil {
			if err := g.HTTP.setPassword(g.Config.Password); err != nil {
				return g.errLog(logger, err)
			}
		}

	} else {
		logger = logger.WithField("action", "generate-password")
		logger.Debug()
		generated, err := g.HTTP.generatePassword()
		if err != nil {
			return g.errLog(logger, err)
		}
		g.Config.Password = generated
	}

	return g.HTTP.configureEmail()
}

func (g *Gerrit) setupSSHClient() error {
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "ssh-client",
	})
	logger.Debug()

	client, err := NewSSHClient(g.Config, g.SSHPort)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	g.SSH = client
	return nil
}

// pushConfig pushes configuration data to the Gerrit instance. This ensures
// that certain settings, such as permissions around the Verified +1 tag, are
// set properly.
func (g *Gerrit) pushConfig() error { // nolint: gocyclo
	logger := g.log.WithFields(log.Fields{
		"phase": "setup",
		"task":  "push-config",
	})
	logger.Debug()

	logger.WithField("action", "new-repo").Debug()
	repo, err := NewRepository(g.Config)
	if err != nil {
		return err
	}
	defer repo.Destroy() // nolint: errcheck

	if err := repo.AddOriginFromContainer(g.Container, "All-Projects"); err != nil {
		return err
	}

	if _, _, err := repo.Git([]string{
		"fetch", "origin", "refs/meta/config:refs/remotes/origin/meta/config"}); err != nil {
		return err
	}

	logger.WithField("action", "checkout").Debug()
	if _, _, err := repo.Git([]string{"checkout", "meta/config"}); err != nil {
		return err
	}

	path := filepath.Join(repo.Root, "project.config")
	ini, err := newProjectConfig(path)
	if err != nil {
		return err
	}
	if err := ini.write(path); err != nil {
		return err
	}

	logger.WithField("action", "add").Debug()
	if _, _, err := repo.Git(append(DefaultGitCommands["add"], path)); err != nil {
		return err
	}

	logger.WithField("action", "commit").Debug()
	if _, _, err := repo.Git([]string{"commit", "--message", "add verified label"}); err != nil {
		return err
	}

	logger.WithField("action", "push").Debug()
	_, _, err = repo.Git([]string{"push", "origin", "meta/config:meta/config"})
	return err
}

// CreateChange will return a *Change struct. If a change has already been
// created then that change will be returned instead of creating a new one.
func (g *Gerrit) CreateChange(project string, subject string) (*Change, error) { // nolint: gocyclo
	logger := g.log.WithField("phase", "create-change")
	client, err := g.HTTP.Gerrit()
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	if project == "" {
		project = ProjectName
	}

	logger = logger.WithField("project", project)
	logger.Debug()

	// Create the project if it does not already exist.
	if _, response, err := client.Projects.GetProject(project); err != nil {
		if response.StatusCode == http.StatusNotFound {
			logger.WithField("action", "create-project").Debug()
			if _, _, err := client.Projects.CreateProject(project, nil); err != nil {
				return nil, err
			}
		}
	}

	path, err := ioutil.TempDir("", fmt.Sprintf("%s-", ProjectName))
	if err != nil {
		return nil, err
	}

	logger = logger.WithFields(log.Fields{
		"path":   path,
		"action": "new-repo",
	})
	logger.Debug()
	repo, err := NewRepository(g.Config)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithField("action", "add-remote-container").Debug()
	if err := repo.AddOriginFromContainer(g.Container, project); err != nil {
		return nil, err
	}

	logger.WithField("action", "commit").Debug()
	if err := repo.Commit(subject); err != nil {
		return nil, err
	}
	id, err := repo.ChangeID()
	if err != nil {
		return nil, err
	}
	return &Change{
		api: client,
		log: g.log.WithFields(log.Fields{
			"cmp": "change",
			"id":  id,
		}),
		Repo:     repo,
		ChangeID: id,
	}, nil
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
	errs := errset.ErrSet{}
	if g.Config.CleanupContainer && g.Container != nil {
		errs = append(errs, g.Container.Terminate())
	}

	if g.SSH != nil {
		errs = append(errs, g.SSH.Close())
	}

	if g.Config.CleanupPrivateKey && g.Config.PrivateKeyPath != "" {
		errs = append(errs, os.Remove(g.Config.PrivateKeyPath))
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

	g := &Gerrit{
		log:    log.WithField("cmp", "core"),
		Config: cfg,
	}
	if err := g.setupSSHKey(); err != nil {
		return g, err
	}
	if err := g.startContainer(); err != nil {
		return g, err
	}

	if cfg.SkipSetup {
		return g, nil
	}

	if err := g.setupHTTPClient(); err != nil {
		return g, err
	}
	if err := g.setupSSHClient(); err != nil {
		return g, err
	}
	if err := g.pushConfig(); err != nil {
		return g, err
	}

	return g, nil
}

// NewFromJSON reads information from a json file and returns a *Gerrit
// struct.
func NewFromJSON(path string) (*Gerrit, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	g := &Gerrit{
		log: log.WithField("cmp", "core"),
	}
	if err := json.Unmarshal(data, g); err != nil {
		return nil, err
	}
	g.Config.Context = ctx
	g.Container.ctx = ctx

	docker, err := dockertest.NewClient()
	if err != nil {
		return nil, err
	}
	g.Container.Docker = docker

	sshClient, err := NewSSHClient(g.Config, g.SSHPort)
	if err != nil {
		return nil, err
	}
	g.SSH = sshClient

	httpClient, err := NewHTTPClient(g.Config, g.HTTPPort)
	if err != nil {
		return nil, err
	}
	g.HTTP = httpClient

	return g, g.pushConfig()
}
