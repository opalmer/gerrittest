package gerrittest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/opalmer/gerrittest/internal"
)

var (
	// DefaultTempName is used as the prefix or suffix of temporary files
	// and folders.
	DefaultTempName = "gerrittest-"

	// DefaultGitCommands contains a mapping of git commands
	// to their arguments. This is used by Repository for running
	// git commands.
	DefaultGitCommands = map[string][]string{
		"status": {"status", "--porcelain"},
		"init":   {"init", "--quiet"},
		"config": {"config", "--local"},
	}

	// DefaultCommitHookName is the name of the hook installed by
	// InstallCommitHook.
	DefaultCommitHookName = "commit-msg"

	// ErrRepositoryNotInitialized is returned any function that needs an
	// initialized repository.
	ErrRepositoryNotInitialized = errors.New("Repository not initialized")
)

// RepositoryConfig is used to store information about a repository.
type RepositoryConfig struct {
	// Path is the path to the repository on disk.
	Path string

	// Command is the git command to run. Defaults to 'git'
	Command string

	// Ctx is the context to use when running commands. Defaults to
	// context.Background()
	Ctx context.Context

	// CommandTimeout is the amount
	CommandTimeout time.Duration

	// PrivateKey is the path to the private key to use for communicating
	// with the git server. Certain functions may return errors if this value
	// is not set.
	PrivateKey string

	// GitConfig are key:value parts of git configuration options
	// to set. Defaults to:
	// {
	//   "user.name":       "admin",
	//   "user.email":      "admin@localhost",
	//   "core.sshCommand": "ssh -i {PrivateKey} -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no",
	// }
	GitConfig map[string]string
}

// Repository is used to store information about an interact
// with a git repository.
type Repository struct {
	mtx  *sync.Mutex
	init bool
	Cfg  *RepositoryConfig
}

// setEnvironment sets up the environment for the given command.
func (r *Repository) setEnvironment(cmd *exec.Cmd) error {
	// We'll need the user to set $HOME otherwise some git commands won't work.
	usr, err := user.Current()
	if err != nil {
		return err
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", usr.HomeDir))

	// Set environment variables to ensure the proper ssh command is run. Not
	// all versions of git support core.sshCommand from the config.
	for _, key := range []string{"GIT_SSH_COMMAND", "GIT_SSH"} {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, r.Cfg.GitConfig["core.sshCommand"]))
	}
	return nil
}

func (r *Repository) run(cmd *exec.Cmd) (string, string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	if err := cmd.Start(); err != nil {
		return "", "", err
	}
	bytesOut, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", "", err
	}
	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return string(bytesOut), "", err
	}
	return string(bytesOut), string(bytesErr), cmd.Wait()
}

// Git runs git with the provided arguments. This also ensures the proper
// working path and environment are set before calling git.
func (r *Repository) Git(args []string) (string, string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	// Change directories to the directory of the repository. Technically
	// there's a -C flag but not all versions of git have this flag and not
	// all subcommands respect it the same way.
	defer os.Chdir(workdir) // nolint: errcheck
	if err := os.Chdir(r.Cfg.Path); err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(r.Cfg.Ctx, r.Cfg.CommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.Cfg.Command, args...)
	if err := r.setEnvironment(cmd); err != nil {
		return "", "", err
	}
	return r.run(cmd)
}

// Status returns the current status of the repository.
func (r *Repository) Status() (string, error) {
	stdout, _, err := r.Git(DefaultGitCommands["status"])
	return stdout, err
}

// Init calls 'git init' on the repository but only if it does not appear
// to already be a repository.
func (r *Repository) Init() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.init {
		return nil
	}
	if _, err := r.Status(); err != nil {
		if _, _, err := r.Git(DefaultGitCommands["init"]); err != nil {
			return err
		}
	}
	r.init = true
	return nil
}

// InstallCommitHook copies the commit hook into the hooks directory of the
// repository. If the repository has not been initialized yet an error will
// be returned. Note, this function will overwrite the existing commit-msg hook
// by default.
func (r *Repository) InstallCommitHook() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.init {
		return ErrRepositoryNotInitialized
	}
	return ioutil.WriteFile(
		filepath.Join(r.Cfg.Path, ".git", "hooks", DefaultCommitHookName),
		internal.MustAsset("internal/commit-msg"), 0700)
}

// Config will call `git config --local key value`.
func (r *Repository) Config(key string, value string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.init {
		return ErrRepositoryNotInitialized
	}
	_, _, err := r.Git(append(DefaultGitCommands["config"], key, value))
	return err
}

// Configure will iterate over the configuration keys provided
// by the RepositoryConfig struct and call Config() on each.
func (r *Repository) Configure() error {
	for key, value := range r.Cfg.GitConfig {
		if err := r.Config(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Add adds a path to the repository. The path must be relative to the root of
// the repository.
func (r *Repository) Add(path string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.init {
		return ErrRepositoryNotInitialized
	}
	return nil
}

// Commit will add a new commit to the repository with the
// given message.
func (r *Repository) Commit(message string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.init {
		return ErrRepositoryNotInitialized
	}
	return nil
}

// Push will push changes to the given remote and reference. `remote` will
// default to 'origin' if not provided and `ref` will default to
// 'HEAD:refs/for/master' if not provided.
func (r *Repository) Push(remote string, ref string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.init {
		return ErrRepositoryNotInitialized
	}
	return nil
}

// CreateRemoteFromSpec adds a new remote based on the provided spec.
// nolint: unused,gosimple,unconvert,varcheck
func (r *Repository) CreateRemoteFromSpec(service *ServiceSpec, remoteName string, project string) error {
	return nil
}

// Remove will remove the entire repository from disk, useful for temporary
// repositories. This cannot be reversed.
func (r *Repository) Remove() error {
	return os.RemoveAll(r.Cfg.Path)
}

// NewRepository constructs and returns a *Repository struct. It will also
// ensure the repository is properly setup before returning.
func NewRepository(cfg *RepositoryConfig) (*Repository, error) {
	repo := &Repository{
		mtx:  &sync.Mutex{},
		init: false,
		Cfg:  cfg}
	if err := repo.Init(); err != nil {
		return nil, err
	}
	if err := repo.InstallCommitHook(); err != nil {
		return nil, err
	}
	if err := repo.Configure(); err != nil {
		return nil, err
	}

	return repo, nil
}

// NewRepositoryConfig returns a *RepositoryConfig struct. If no path is
// provided then one will be generated for you.
func NewRepositoryConfig(path string, privateKey string) (*RepositoryConfig, error) {
	if privateKey == "" {
		return nil, errors.New("Missing private key")
	}

	if path == "" {
		newPath, err := ioutil.TempDir("", DefaultTempName)
		if err != nil {
			return nil, err
		}
		path = newPath
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}

	return &RepositoryConfig{
		Path:           path,
		Ctx:            context.Background(),
		Command:        "git",
		CommandTimeout: time.Minute * 10,
		PrivateKey:     privateKey,
		GitConfig: map[string]string{
			"user.name":       "admin",
			"user.email":      "admin@localhost",
			"core.sshCommand": fmt.Sprintf("ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", privateKey),
		},
	}, nil
}
