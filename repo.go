package gerrittest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"time"

	"github.com/opalmer/gerrittest/internal"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	// DefaultTempName is used as the prefix or suffix of temporary files
	// and folders.
	DefaultTempName = "gerrittest-"

	// ErrRepositoryNotInitialized may be returned by any operation that
	// requires a fully setup Repository struct.
	ErrRepositoryNotInitialized = errors.New("The repository is not initialized")
)

// Repository is used to store information about an interact
// with a git repository.
type Repository struct {
	Path   string
	Repo   *git.Repository
	User   string // Defaults to 'admin' in Init()
	Email  string // Defaults to '<User>@localhost' in Init()
	Branch string // Defaults to 'master' in Init()
}

func (r *Repository) writeConfig() error {
	if r.Repo == nil {
		return ErrRepositoryNotInitialized
	}
	cfg, err := r.Repo.Config()
	if err != nil {
		return err
	}
	data, err := cfg.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(r.Path, ".git", "config"), data, 0600)
}

// CreateRemoteFromSpec adds a new remote based on the provided spec.
func (r *Repository) CreateRemoteFromSpec(service *ServiceSpec, remoteName string, project string) error {
	_, err := r.Repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URL: fmt.Sprintf(
			"ssh://%s@%s:%d/%s", service.Admin.Login,
			service.SSH.Address, service.SSH.Public, project),
		Fetch: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
	})
	if err != nil {
		return err
	}
	return r.writeConfig()
}

// Init initializes the git repository. If the repository was setup without
// a path then a temp path will be used. Note, this may make modifications to
// an existing repository.
func (r *Repository) Init() error {
	if r.Repo != nil {
		return nil
	}

	// Configure defaults.
	if r.User == "" {
		r.User = "admin"
	}
	if r.Email == "" {
		r.Email = fmt.Sprintf("%s@localhost", r.User)
	}
	if r.Branch == "" {
		r.Branch = "master"
	}
	if r.Path == "" {
		path, err := ioutil.TempDir("", DefaultTempName)
		if err != nil {
			return err
		}
		r.Path = path
	}

	// Create the repository.
	if _, err := os.Stat(filepath.Join(r.Path, ".git")); os.IsNotExist(err) {
		repo, err := git.PlainInit(r.Path, false)
		if err != nil {
			return err
		}
		r.Repo = repo

		// Open an existing repository.
	} else {
		repo, err := git.PlainOpen(r.Path)
		if err != nil {
			return err
		}
		r.Repo = repo
	}

	// Drop the commit hook on disk.
	os.MkdirAll(filepath.Join(r.Path, ".git", "hooks"), 0700)
	if err := ioutil.WriteFile(
		filepath.Join(r.Path, ".git", "hooks", "commit-msg"),
		internal.MustAsset("internal/commit-msg"), 0700); err != nil {
		return err
	}

	// Add user/email to the config and write it to disk.
	cfg, err := r.Repo.Config()
	if err != nil {
		return err
	}
	cfg.Raw = cfg.Raw.AddOption("core", "", "user", r.User)
	cfg.Raw = cfg.Raw.AddOption("core", "", "email", r.Email)
	return r.writeConfig()
}

// Add adds a path to the repository. The path must be relative to the root of
// the repository
func (r *Repository) Add(path string) error {
	if r.Repo == nil {
		return ErrRepositoryNotInitialized
	}

	tree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	_, err = tree.Add(path)
	return err
}

// Commit will add a new commit to the repository with the
// given message.
func (r *Repository) Commit(message string) error {
	if r.Repo == nil {
		return ErrRepositoryNotInitialized
	}

	tree, err := r.Repo.Worktree()
	if err != nil {
		return err
	}
	author := &object.Signature{
		Name:  r.User,
		Email: r.Email,
		When:  time.Now(),
	}
	_, err = tree.Commit(message, &git.CommitOptions{
		All:       false,
		Author:    author,
		Committer: author,
	})
	return err
}

// Remove will remove the entire repository from disk, useful for temporary
// repositories. This cannot be reversed.
func (r *Repository) Remove() error {
	if r.Path == "" {
		return ErrRepositoryNotInitialized
	}
	return os.RemoveAll(r.Path)
}
