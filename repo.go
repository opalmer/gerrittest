package gerrittest

import (
	"bytes"
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

	"github.com/opalmer/gerrittest/internal"
	log "github.com/sirupsen/logrus"
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
		"commit":              {"commit", "--allow-empty", "--message"},
		"push":                {"push", "--porcelain"},
		"last-commit-message": {"log", "-n", "1", "--format=medium"},
		"amend":               {"commit", "--amend", "--no-edit", "--allow-empty"},
	}

	// DefaultCommitHookName is the name of the hook installed by
	// installCommitHook.
	DefaultCommitHookName = "commit-msg"

	// ErrRemoteDoesNotExist is returned by GetRemoteURL if the requested
	// remote does not appear to exist.
	ErrRemoteDoesNotExist = errors.New("requested remote does not exist")

	// ErrRemoteExists is returned by AddRemote if the provided remote already
	// exists.
	ErrRemoteExists = errors.New("remote with the given name already exists")

	//ErrFailedToLocateChange is returned by functions, such as Push(), that
	// expect to find a change number in the output from git.
	ErrFailedToLocateChange = errors.New("failed to locate ChangeID")

	// ErrRemoteNotProvided is returned by functions when a remote name
	// or value is required by not provided.
	ErrRemoteNotProvided = errors.New("remote not provided")

	// ErrNoCommits is returned by ChangeID if there are not any commits
	// to the repository yet.
	ErrNoCommits = errors.New("no commits")

	// RegexChangeID is used to match the Change-Id for a commit.
	RegexChangeID = regexp.MustCompile(`(?m)^\s+Change-Id: (I[a-f0-9]{40}).*$`)
)

// Diff is a struct which represents a single commit to the
// repository.
type Diff struct {
	Error   error
	Content []byte
	Commit  string
}

// Repository is used to store information about an interact
// with a git repository. In the end, this is a thin wrapper
// around GitConfig commands.
type Repository struct {
	config *Config
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, r.config.GitConfig["core.sshCommand"]))
	}
	return nil
}

func (r *Repository) run(cmd *exec.Cmd) (string, string, error) {
	logger := log.WithFields(log.Fields{
		"phase": "git",
		"repo":  r.config.PrivateKeyPath,
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
	if err := os.Chdir(r.config.RepoRoot); err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(r.config.Context, r.config.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.config.GitCommand, args...)
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
	absolute := filepath.Join(r.config.RepoRoot, path)
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

// Push will push changes to the given remote and reference. `ref`
// will default to 'HEAD:refs/for/master' if not provided.
func (r *Repository) Push(remote string, ref string) error {
	if remote == "" {
		return ErrRemoteNotProvided
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
		return ErrRemoteNotProvided
	}

	return r.AddRemote(remote, fmt.Sprintf(
		"ssh://%s@%s:%d/%s", r.config.GitConfig["user.name"],
		container.SSH.Address, container.SSH.Public, project))
}

// ChangeID returns a string representing the change id of the last
// commit.
func (r *Repository) ChangeID() (string, error) {
	stdout, stderr, err := r.Git(DefaultGitCommands["last-commit-message"])
	if strings.Contains(stderr, "does not have any commits yet") {
		return "", ErrNoCommits
	}

	if err != nil {
		return "", err
	}

	matches := RegexChangeID.FindAllStringSubmatch(stdout, -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1], nil
	}
	return "", ErrFailedToLocateChange
}

// Amend amends the current commit.
func (r *Repository) Amend() error {
	_, _, err := r.Git(DefaultGitCommands["amend"])
	return err
}

// PlaybackFrom will play changes back from the given source into this
// repository.
func (r *Repository) PlaybackFrom(source PlaybackSource) error {
	diffs, err := source.Read(r.config.Context)
	if err != nil {
		return err
	}

	failures := 0
	logger := log.WithFields(log.Fields{
		"cmp":    "repo",
		"phase":  "playback",
		"action": "apply",
	})

	// TODO make sure we only do this once.
	if err := r.AddContent(".empty", 0600, []byte("")); err != nil {
		return err
	}

	if err := r.Push("", ""); err != nil {
		return err
	}

	for diff := range diffs {
		logger = logger.WithFields(log.Fields{
			"failures": failures,
		})
		logger.Debug()

		cmd := exec.Command("git", "-C", r.config.RepoRoot, "apply")
		cmd.Stdin = bytes.NewBuffer(diff.Content)
		if err := cmd.Run(); err != nil {
			logger.WithError(err).Warn()
			failures++
			continue
		}

		if err := r.Amend(); err != nil {
			logger.WithError(err).Warn()
			failures++
			continue
		}
		if err := r.Push("", ""); err != nil {
			logger.WithError(err).Warn()
			failures++
			continue
		}
	}

	if failures > 0 {
		return errors.New("failed to apply any diffs")
	}

	return nil
}

// Remove will remove the entire repository from disk, useful for temporary
// repositories. This cannot be reversed.
func (r *Repository) Remove() error {
	if r.config.CleanupGitRepo {
		return os.RemoveAll(r.config.RepoRoot)
	}
	return nil
}

// WriteCommitHook writes the default commit hook to the provided repository.
func WriteCommitHook(repo *Repository, hookName string) error {
	return ioutil.WriteFile(
		filepath.Join(repo.config.RepoRoot, ".git", "hooks", hookName),
		internal.MustAsset("internal/commit-msg"), 0700)
}

// NewRepository constructs and returns a *Repository struct. It will also
// ensure the repository is properly setup before returning.
func NewRepository(config *Config) (*Repository, error) {
	if config.PrivateKeyPath == "" {
		return nil, errors.New("missing private key")
	}

	if config.RepoRoot == "" {
		config.CleanupGitRepo = true
		tmppath, err := ioutil.TempDir("", fmt.Sprintf("%s-", ProjectName))
		if err != nil {
			return nil, err
		}
		config.RepoRoot = tmppath
	}

	if err := os.MkdirAll(config.RepoRoot, 0700); err != nil {
		return nil, err
	}
	repo := &Repository{config: config}

	if err := repo.Init(); err != nil {
		return nil, err
	}

	if err := WriteCommitHook(repo, DefaultCommitHookName); err != nil {
		return nil, err
	}

	for key, value := range repo.config.GitConfig {
		if err := repo.ConfigLocal(key, value); err != nil {
			return nil, err
		}
	}

	return repo, nil
}
