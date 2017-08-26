package gerrittest

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fmt"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/gerrittest/internal"
)

var (
	// CommandTimeout is the default amount of time to wait
	// for a git command to finish.
	CommandTimeout = time.Second * 60

	// DefaultDirectoryMode is used when creating parent directories
	// for files.
	DefaultDirectoryMode os.FileMode = 0700

	// DefaultFileMode is returned by FileInput.GetMode if no mode
	// has been set.
	DefaultFileMode os.FileMode = 0600
)

// FileInput defines all the information necessary to add a single file to
// a repository.
type FileInput struct {
	Path    string
	Content []byte
	Mode    os.FileMode
}

// GetMode returns the mode of the file. If no mode has been set `0600` will
// be returned.
func (f *FileInput) GetMode() os.FileMode {
	if f.Mode == 0 {
		return DefaultFileMode
	}
	return f.Mode
}

// Repository is a basic wrapper for a git repository. This does not
// fully implement all git commands, just enough to work with
// a repository on disk for the purposes of gerrittest.
type Repository struct {
	Root       string
	Git        string
	mtx        *sync.Mutex
	log        *log.Entry
	sshCommand string
}

// Run will run git with the provided args to completion then return the
// results. The working directory will be changed to the root of the repository
// prior to running and be reset on exit.
func (r *Repository) Run(args []string) (string, string, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	logger := r.log.WithFields(log.Fields{
		"action": "run",
		"args":   args,
	})
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	if err := os.Chdir(r.Root); err != nil {
		return "", "", err
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, r.Git, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Depending on the version of Git either the config command in
	// ConfigureFromSpec() will work or one of the below will.
	cmd.Env = append(
		cmd.Env, fmt.Sprintf("GIT_SSH_COMMAND=%s", r.sshCommand))
	cmd.Env = append(
		cmd.Env, fmt.Sprintf("GIT_SSH=%s", r.sshCommand))

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)
	outstr := stdout.String()
	errstr := stderr.String()
	logger = logger.WithFields(log.Fields{
		"duration": duration,
		"stdout":   outstr,
		"stderr":   errstr,
	})

	if err != nil {
		logger.Warn()
		defer os.Chdir(cwd)
		return outstr, errstr, err
	}

	logger.Debug()
	return outstr, errstr, os.Chdir(cwd)
}

// Init will create the repository if it does not exist. This function
// does nothing if the path on disk already appears to be a repository.
func (r *Repository) Init() error {
	if _, err := os.Stat(r.Root); os.IsNotExist(err) {
		if err := os.MkdirAll(r.Root, DefaultDirectoryMode); err != nil {
			return err
		}
	}

	r.log.WithFields(log.Fields{
		"action": "init",
		"path":   r.Root,
	}).Debug()
	_, _, err := r.Run([]string{"init", "--quiet", r.Root})
	config := map[string]string{
		"user.name":  "admin",
		"user.email": "admin@localhost",
	}
	if err := r.ConfigureValues(config); err != nil {
		return err
	}

	hook, err := internal.Asset("internal/commit-msg")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(
		filepath.Join(r.Root, ".git", "hooks", "commit-msg"), hook, 0700)
}

// Destroy removes the repository from the local disk.
func (r *Repository) Destroy() error {
	r.log.WithFields(log.Fields{
		"action": "destroy",
		"path":   r.Root,
	}).Debug()
	return os.RemoveAll(r.Root)
}

// Add adds a single file to the repository.
func (r *Repository) Add(input *FileInput) error {
	r.log.WithFields(log.Fields{
		"action": "add",
		"path":   input.Path,
	}).Debug()
	_, _, err := r.Run([]string{"add", input.Path})
	return err
}

// AddFile performs multiple steps in a single command:
//  - Write a file to disk using the given path. The path itself should
//    be relative to the repository root. Any parent paths will be automatically
//    created.
//  - Set the permissions of the file.
//  - Add the file to the git repository.
func (r *Repository) AddFile(input *FileInput) error {
	logger := r.log.WithFields(log.Fields{
		"action": "add-file",
		"path":   input.Path,
	})

	absolute := filepath.Join(r.Root, input.Path)
	logger = logger.WithField("phase", "mkdir")
	if err := os.MkdirAll(filepath.Dir(absolute), DefaultDirectoryMode); err != nil {
		logger.WithError(err).Warn()
		return err
	}

	logger = logger.WithField("phase", "write")
	if err := ioutil.WriteFile(absolute, input.Content, input.GetMode()); err != nil {
		logger.WithError(err).Warn()
		return err
	}
	return r.Add(input)
}

// Commit will commit any pending changes with the provided message.
func (r *Repository) Commit(message string) error {
	r.log.WithFields(log.Fields{
		"action":  "commit",
		"message": message,
	}).Debug()
	_, _, err := r.Run([]string{"commit", "-m", message, "--quiet"})
	return err
}

// Amend amends the current commit without changing the commit message.
func (r *Repository) Amend() error {
	r.log.WithFields(log.Fields{
		"action": "amend",
	}).Debug()
	_, _, err := r.Run(
		[]string{"commit", "--quiet", "--amend", "--no-edit", "--allow-empty"})
	return err
}

// Push will call 'git push HEAD:refs/for/<branch>'. Note, this will
// return an error if you have not already configured the repository. If no
// branch is provided 'master' will be used.
func (r *Repository) Push(branch string) error {
	if branch == "" {
		branch = "master"
	}
	r.log.WithFields(log.Fields{
		"action": "amend",
	}).Debug()
	_, _, err := r.Run(
		[]string{"push", "origin", fmt.Sprintf("HEAD:refs/for/%s", branch)})
	return err
}

// ConfigureValues runs 'git config' using the given map.
func (r *Repository) ConfigureValues(values map[string]string) error {
	logger := r.log.WithFields(log.Fields{
		"action": "config",
	})
	for key, value := range values {
		logger.WithFields(log.Fields{
			"action": "config",
			"key":    key,
			"value":  value,
		}).Debug()
		if _, _, err := r.Run(
			[]string{"config", "--local", key, value}); err != nil {
			logger.WithError(err).Error()
			return err
		}
	}
	return nil
}

// ConfigureFromSpec will configure the git repository to point at
func (r *Repository) ConfigureFromSpec(service *ServiceSpec, project string, branch string) error {
	r.mtx.Lock()
	r.sshCommand = fmt.Sprintf(
		"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i %s",
		service.Admin.PrivateKey)
	config := map[string]string{
		"remote.origin.url": fmt.Sprintf(
			"ssh://%s@%s:%d/%s", service.Admin.Login,
			service.SSH.Address, service.SSH.Public, project),
		"remote.origin.fetch":                   "+refs/heads/*:refs/remotes/origin/*",
		fmt.Sprintf("branch.%s.remote", branch): "origin",
		fmt.Sprintf("branch.%s.merge", branch):  fmt.Sprintf("refs/heads/%s", branch),
		"core.sshCommand":                       r.sshCommand,
	}
	r.mtx.Unlock()
	return r.ConfigureValues(config)
}

// NewRepository creates and returns a *Repository struct. If root is defined
// as "" then a temporary will be created.
func NewRepository(root string) (*Repository, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	if root == "" {
		path, err := ioutil.TempDir("", "gerrittest-")
		if err != nil {
			return nil, err
		}
		root = path
	}

	repo := &Repository{
		Root: root,
		Git:  git,
		mtx:  &sync.Mutex{},
		log:  log.WithField("cmp", "repo"),
	}
	return repo, repo.Init()
}
