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
	"regexp"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/gerrittest/internal"
)

var (
	// DefaultGitCommands contains a mapping of git commands
	// to their arguments. This is used by Repository for running
	// git commands.
	DefaultGitCommands = map[string][]string{
		"status":              {"status", "--porcelain"},
		"init":                {"init", "--quiet"},
		"config":              {"config", "--local"},
		"add":                 {"add", "--force"},
		"remote-add":          {"remote", "add"},
		"get-remote-url":      {"remote", "get-url"},
		"commit":              {"commit", "--message"},
		"push":                {"push", "--porcelain"},
		"last-commit-message": {"log", "-n", "1", "--format=medium"},
	}

	// DefaultCommitHookName is the name of the hook installed by
	// installCommitHook.
	DefaultCommitHookName = "commit-msg"

	// ErrRemoteDoesNotExist is returned by GetRemoteURL if the requested
	// remote does not appear to exist.
	ErrRemoteDoesNotExist = errors.New("Requested remote does not exist")

	// ErrRemoteExists is returned by AddRemote if the provided remote already
	// exists.
	ErrRemoteExists = errors.New("Remote with the given name already exists")

	//ErrFailedToLocateChange is returned by functions, such as Push(), that
	// expect to find a change number in the output from git.
	ErrFailedToLocateChange = errors.New("Failed to locate ChangeID")

	// RegexChangeID is used to match the Change-Id for a commit.
	RegexChangeID = regexp.MustCompile(`(?m)^\s+Change-Id: (I[a-f0-9]{40}).*$`)
)

// RepositoryConfig is used to store information about a repository.
type RepositoryConfig struct {
	// Path is the path to the repository on disk.
	Path string `json:"path"`

	// Command is the git command to run. Defaults to 'git'
	Command string `json:"command"`

	// Ctx is the context to use when running commands. Defaults to
	// context.Background()
	Ctx context.Context `json:"-"`

	// CommandTimeout is the amount
	CommandTimeout time.Duration `json:"command_timeout"`

	// PrivateKeyPath is the path to the private key to use for communicating
	// with the git server. Certain functions may return errors if this value
	// is not set.
	PrivateKeyPath string `json:"private_key_path"`

	// GitConfig are key:value parts of git configuration options
	// to set. Defaults to:
	// {
	//   "user.name":       "admin",
	//   "user.email":      "admin@localhost",
	//   "core.sshCommand": "ssh -i {PrivateKeyPath} -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no",
	// }
	GitConfig map[string]string `json:"git_config"`
}

// newRepositoryConfig returns a *RepositoryConfig struct. If no path is
// provided then one will be generated for you.
func newRepositoryConfig(path string, privateKey string) (*RepositoryConfig, error) {
	if privateKey == "" {
		return nil, errors.New("Missing private key")
	}

	if path == "" {
		tmppath, err := ioutil.TempDir("", "gerritest-")
		if err != nil {
			return nil, err
		}
		path = tmppath
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}

	return &RepositoryConfig{
		Path:           path,
		Ctx:            context.Background(),
		Command:        "git",
		CommandTimeout: time.Minute * 10,
		PrivateKeyPath: privateKey,
		GitConfig: map[string]string{
			"user.name":       "admin",
			"user.email":      "admin@localhost",
			"core.sshCommand": fmt.Sprintf("ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", privateKey),
		},
	}, nil
}

// Repository is used to store information about an interact
// with a git repository. In the end, this is a thin wrapper
// around Git commands.
type Repository struct {
	Config *RepositoryConfig
	Path   string `json:"path"`
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, r.Config.GitConfig["core.sshCommand"]))
	}
	return nil
}

func (r *Repository) run(cmd *exec.Cmd) (string, string, error) {
	logger := log.WithFields(log.Fields{
		"phase": "git",
		"repo":  r.Path,
		"cmd":   strings.Join(cmd.Args, " "),
	})
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	if err := cmd.Start(); err != nil {
		logger.WithError(err).Error()
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
	err = cmd.Wait()
	sOut := string(bytesOut)
	sErr := string(bytesErr)
	code := 0
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			code = status.ExitStatus()
		}
	}
	logger.WithFields(log.Fields{
		"stdout": sOut,
		"stderr": sErr,
		"code":   code,
	}).Debug()
	return string(bytesOut), string(bytesErr), err
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
	if err := os.Chdir(r.Path); err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(r.Config.Ctx, r.Config.CommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.Config.Command, args...)
	if err := r.setEnvironment(cmd); err != nil {
		return "", "", err
	}
	return r.run(cmd)
}

// Init calls 'git init' on the repository root but only if 'git status'
// fails.
func (r *Repository) Init() error {
	if _, err := r.Status(); err == nil {
		return nil
	}
	_, _, err := r.Git(DefaultGitCommands["init"])
	return err
}

// Status returns the current status of the repository.
func (r *Repository) Status() (string, error) {
	stdout, _, err := r.Git(DefaultGitCommands["status"])
	return stdout, err
}

// ConfigLocal will call `git config --local key value`.
func (r *Repository) ConfigLocal(key string, value string) error {
	_, _, err := r.Git(append(DefaultGitCommands["config"], key, value))
	return err
}

// Add adds a path to the repository. The path must be relative to the root of
// the repository.
func (r *Repository) Add(paths ...string) error {
	_, _, err := r.Git(append(DefaultGitCommands["add"], paths...))
	return err
}

// AddContent is similar to Add() except it allows content to be created in
// addition to be added to the repo.
func (r *Repository) AddContent(path string, mode os.FileMode, content []byte) error {
	absolute := filepath.Join(r.Path, path)
	if err := os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(absolute, content, mode); err != nil {
		return err
	}
	return r.Add(path)
}

// Commit will add a new commit to the repository with the
// given message.
func (r *Repository) Commit(message string) error {
	_, _, err := r.Git(append(DefaultGitCommands["commit"], message))
	return err
}

// Push will push changes to the given remote and reference. `remote` will
// default to 'origin' if not provided and `ref` will default to
// 'HEAD:refs/for/master' if not provided. The returned ingeter will
func (r *Repository) Push(remote string, ref string) error {
	if remote == "" {
		remote = "origin"
	}
	if ref == "" {
		ref = "HEAD:refs/for/master"
	}

	_, _, err := r.Git(append(DefaultGitCommands["push"], remote, ref))
	return err
}

// GetRemoteURL will return the url for the given remote name. If the requested
// remote does not exist ErrRemoteDoesNotExist will be returned.
func (r *Repository) GetRemoteURL(name string) (string, error) {
	stdout, stderr, err := r.Git(append(DefaultGitCommands["get-remote-url"], name))
	if err != nil && strings.Contains(stderr, "No such remote") {
		return "", ErrRemoteDoesNotExist
	}
	return strings.TrimSpace(stdout), nil
}

// AddRemote will add a remote with the given name so long as it does not
// already exist.
func (r *Repository) AddRemote(name string, uri string) error {
	if _, err := r.GetRemoteURL(name); err == nil {
		return ErrRemoteExists
	}
	_, _, err := r.Git(append(DefaultGitCommands["remote-add"], name, uri))
	return err
}

// AddRemoteFromContainer adds a new remote based on the provided container.
func (r *Repository) AddRemoteFromContainer(container *Container, remote string, project string) error {
	if remote == "" {
		remote = "origin"
	}

	return r.AddRemote(remote, fmt.Sprintf(
		"ssh://%s@%s:%d/%s", r.Config.GitConfig["user.name"],
		container.SSH.Address, container.SSH.Public, project))
}

// ChangeID returns a string representing the change id of the last
// commit.
func (r *Repository) ChangeID() (string, error) {
	stdout, _, err := r.Git(DefaultGitCommands["last-commit-message"])
	if err != nil {
		return "", err
	}

	matches := RegexChangeID.FindAllStringSubmatch(stdout, -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1], nil
	}
	return "", ErrFailedToLocateChange
}

// Remove will remove the entire repository from disk, useful for temporary
// repositories. This cannot be reversed.
func (r *Repository) Remove() error {
	return os.RemoveAll(r.Path)
}

// WriteCommitHook writes the default commit hook to the provided repository.
func WriteCommitHook(repo *Repository, hookName string) error {
	return ioutil.WriteFile(
		filepath.Join(repo.Path, ".git", "hooks", hookName),
		internal.MustAsset("internal/commit-msg"), 0700)
}

// NewRepository constructs and returns a *Repository struct. It will also
// ensure the repository is properly setup before returning.
func NewRepository(cfg *RepositoryConfig) (*Repository, error) {
	repo := &Repository{Config: cfg, Path: cfg.Path}

	if err := repo.Init(); err != nil {
		return nil, err
	}

	if err := WriteCommitHook(repo, DefaultCommitHookName); err != nil {
		return nil, err
	}

	for key, value := range repo.Config.GitConfig {
		if err := repo.ConfigLocal(key, value); err != nil {
			return nil, err
		}
	}

	return repo, nil
}
