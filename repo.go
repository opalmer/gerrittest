package gerrittest

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

var (
	// CommandTimeout is the default amount of time to wait
	// for a git command to finish.
	CommandTimeout = time.Second * 60
)

// Repository is a basic wrapper for a git repository. This does not
// fully implement all git commands, just enough to work with
// a repository on disk for the purposes of gerrittest.
type Repository struct {
	Root string
	Git  string
	log  *log.Entry
}

// Run will run git with the provided args to completion then return the
// results. The working directory will be changed to the root of the repository
// prior to running and be reset on exit.
func (r *Repository) Run(args []string) (string, string, error) {
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
	start := time.Now()
	logger = logger.WithFields(log.Fields{
		"duration": time.Since(start),
	})

	if err := cmd.Run(); err != nil {
		logger = logger.WithError(err)
		logger.Warn()
		defer os.Chdir(cwd)
		return stdout.String(), stderr.String(), err
	}

	logger.Debug()
	return stdout.String(), stderr.String(), os.Chdir(cwd)
}

// Init will create the repository if it does not exist. This function
// does nothing if the path on disk already appears to be a repository.
func (r *Repository) Init() error {
	if _, err := os.Stat(r.Root); os.IsNotExist(err) {
		if err := os.MkdirAll(r.Root, 0700); err != nil {
			return err
		}
	}

	r.log.WithFields(log.Fields{
		"action": "init",
		"path":   r.Root,
	}).Debug()
	_, _, err := r.Run([]string{"init", "--quiet", r.Root})
	return err
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
func (r *Repository) Add(relative string) error {
	r.log.WithFields(log.Fields{
		"action": "add",
		"path":   relative,
	}).Debug()
	_, _, err := r.Run([]string{"add", relative})
	return err
}

// AddFile performs multiple steps in a single command:
//  - Write a file to disk using the given path. The path itself should
//    be relative to the repository root. Any parent paths will be automatically
//    created.
//  - Set the permissions of the file.
//  - Add the file to the git repository.
func (r *Repository) AddFile(relative string, content []byte, mode os.FileMode) error {
	logger := r.log.WithFields(log.Fields{
		"action": "add-file",
		"path":   relative,
	})
	absolute := filepath.Join(r.Root, relative)
	if err := os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		logger.WithError(err).Warn()
		return err
	}
	if err := ioutil.WriteFile(absolute, content, mode); err != nil {
		logger.WithError(err).Warn()
		return err
	}
	return r.Add(relative)
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
		log:  log.WithField("cmp", "repo"),
	}
	return repo, repo.Init()
}
