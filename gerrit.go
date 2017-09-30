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
)

// ProjectName is used anywhere we need a default value (temp files, default
// field values, etc.
const ProjectName = "gerrittest"

// Gerrit is the central struct which combines multiple components
// of the gerrittest project. Use New() to construct this struct.
type Gerrit struct {
	ctx       context.Context
	cancel    context.CancelFunc
	log       *log.Entry
	Config    *Config          `json:"config"`
	Container *Container       `json:"container"`
	HTTP      *HTTPClient      `json:"-"`
	HTTPPort  *dockertest.Port `json:"http"`
	SSH       *SSHClient       `json:"-"`
	SSHPort   *dockertest.Port `json:"ssh"`
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
		g.ctx, g.Config.PortHTTP, g.Config.PortSSH, g.Config.Image)
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

	// If no keys have been provided generate one and add it
	// to the config.
	if len(g.Config.SSHKeys) == 0 {
		key, err := NewSSHKey()
		if err != nil {
			return err
		}
		g.Config.SSHKeys = append(g.Config.SSHKeys, key)
	}

	for _, key := range g.Config.SSHKeys {
		if key.Default {
			g.Config.GitConfig["core.sshCommand"] = fmt.Sprintf(
				"ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", key.Path)
			break
		}
	}

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

	logger.WithField("action", "insert-keys").Debug()
	if err := g.HTTP.insertPublicKeys(); err != nil {
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
	defer gg.cancel()
	errs := errset.ErrSet{}
	if g.Config.CleanupContainer && g.Container != nil {
		errs = append(errs, g.Container.Terminate())
	}

	if g.SSH != nil {
		errs = append(errs, g.SSH.Close())
	}

	for _, key := range g.Config.SSHKeys {
		errs = append(errs, key.Remove())
	}

	return errs.ReturnValue()
}

// New constructs and returns a *Gerrit struct after all setup steps have
// been completed. Once this function returns Gerrit will be running in
// a container, an admin user will be created and a git repository will
// be setup pointing at the service in the container.
func New(cfg *Config) (*Gerrit, error) {
	ctx, cancel := context.WithCancel(cfg.Context)
	g := &Gerrit{
		ctx:    ctx,
		cancel: cancel,
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

// LoadJSON strictly loads the json file from the provided path. It makes no
// attempts to verify that the docker container is running orr
func LoadJSON(path string) (*Gerrit, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	g := &Gerrit{log: log.WithField("cmp", "core")}
	return g, json.Unmarshal(data, g)
}

// NewFromJSON reads information from a json file and returns a *Gerrit
// struct.
func NewFromJSON(path string) (*Gerrit, error) {
	logger := log.WithField("phase", "new-from-json")
	logger.WithFields(log.Fields{
		"path":   path,
		"action": "read",
	}).Debug()
	g, err := LoadJSON(path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx
	g.cancel = cancel

	logger.WithFields(log.Fields{
		"path":   path,
		"action": "get-dockertest-client",
	}).Debug()
	docker, err := dockertest.NewClient()
	if err != nil {
		return nil, err
	}
	g.Container.Docker = docker

	logger.WithFields(log.Fields{
		"path":   path,
		"action": "load-ssh-keys",
	}).Debug()
	for _, key := range g.Config.SSHKeys {
		if err := key.load(); err != nil {
			return nil, err
		}
	}

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
