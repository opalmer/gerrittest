package gerrittest

import (
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
	// GitCommand is the command used to run git commands.
	GitCommand = "git"

	// DefaultGitCommands contains a mapping of git commands
	// to their arguments. This is used by Repository for running
	// git commands.
	DefaultGitCommands = map[string][]string{
		"status":              {"status", "--porcelain"},
		"config":              {"config", "--local"},
		"add":                 {"add", "--force"},
		"remote-add":          {"remote", "add"},
		"get-remote-url":      {"remote", "get-url"},
		"commit":              {"commit", "--allow-empty", "--message"},
		"push":                {"push", "--porcelain"},
		"last-commit-message": {"log", "-n", "1", "--format=medium"},
		"amend":               {"commit", "--amend", "--no-edit", "--allow-empty"},
	}

	// ErrRemoteDoesNotExist is returned by GetRemoteURL if the requested
	// remote does not appear to exist.
	ErrRemoteDoesNotExist = errors.New("requested remote does not exist")

	//ErrFailedToLocateChange is returned by functions, such as Push(), that
	// expect to find a change number in the output from git.
	ErrFailedToLocateChange = errors.New("failed to locate ChangeID")

	// ErrNoCommits is returned by ChangeID if there are not any commits
	// to the repository yet.
	ErrNoCommits = errors.New("no commits")

	// RegexChangeID is used to match the Change-Id for a commit.
	RegexChangeID = regexp.MustCompile(`(?m)^\s+Change-Id: (I[a-f0-9]{40}).*$`)
)

// Repository is used to store information about an interact
// with a git repository. In the end, this is a thin wrapper
// around GitConfig commands.
type Repository struct {
	SSHCommand string
	Root       string
	Username   string
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, r.SSHCommand))
	}
	return nil
}

func (r *Repository) run(cmd *exec.Cmd) (string, string, error) {
	cwd, err := os.Getwd()
	logger := log.WithFields(log.Fields{
		"phase": "run",
		"cmd":   strings.Join(cmd.Args, " "),
		"wd":    cwd,
	})
	if err != nil {
		return "", "", err
	}
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
	if err := os.Chdir(r.Root); err != nil {
		return "", "", err
	}

	cmd := exec.Command(GitCommand, args...)
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

// Add is similar to Add() except it allows content to be created in
// addition to be added to the repo.
func (r *Repository) Add(path string, mode os.FileMode, content []byte) error {
	absolute := filepath.Join(r.Root, path)
	if err := os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(absolute, content, mode); err != nil {
		return err
	}

	_, _, err := r.Git(append(DefaultGitCommands["add"], path))
	return err
}

// Remove removes the requested content from the repository.
func (r *Repository) Remove(path string) error {
	absolute := filepath.Join(r.Root, path)
	stat, err := os.Stat(absolute)
	if os.IsNotExist(err) {
		return nil
	}
	args := []string{"rm"}
	if stat.IsDir() {
		args = append(args, "-r")
	}
	args = append(args, path)
	_, _, err = r.Git(args)
	return err
}

// Commit will add a new commit to the repository with the
// given message.
func (r *Repository) Commit(message string) error {
	_, _, err := r.Git(append(DefaultGitCommands["commit"], message))
	return err
}

// Push will push changes to the given remote and reference. `ref`
// will default to 'HEAD:refs/for/master' if not provided.
func (r *Repository) Push(ref string) error {
	if ref == "" {
		ref = "HEAD:refs/for/master"
	}

	_, _, err := r.Git(append(DefaultGitCommands["push"], "origin", ref))
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
		return nil
	}
	_, _, err := r.Git(append(DefaultGitCommands["remote-add"], name, uri))
	return err
}

// AddOriginFromContainer adds a new remote based on the provided container.
func (r *Repository) AddOriginFromContainer(container *Container, project string) error {
	return r.AddRemote("origin", fmt.Sprintf(
		"ssh://%s@%s:%d/%s", r.Username,
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
	_, stderr, err := r.Git(DefaultGitCommands["amend"])
	if strings.Contains(stderr, "You have nothing to amend") {
		err = ErrNoCommits
	}
	return err
}

// Destroy will remove the entire repository from disk, useful for temporary
// repositories. This cannot be reversed.
func (r *Repository) Destroy() error {
	return os.RemoveAll(r.Root)
}

// NewRepository constructs and returns a *Repository struct. It will also
// ensure the repository is properly setup before returning.
func NewRepository(config *Config) (*Repository, error) {
	root, err := ioutil.TempDir("", fmt.Sprintf("%s-", ProjectName))
	if err != nil {
		return nil, err
	}

	repo := &Repository{
		Username:   config.GitConfig["user.name"],
		SSHCommand: config.GitConfig["core.sshCommand"],
		Root:       root,
	}

	if _, _, err := repo.Git([]string{"init", "--quiet"}); err != nil {
		repo.Destroy() // nolint: errcheck
		return nil, err
	}

	if err := ioutil.WriteFile(
		filepath.Join(repo.Root, ".git", "hooks", "commit-msg"),
		internal.MustAsset("internal/commit-msg"), 0700); err != nil {
		repo.Destroy() // nolint: errcheck
		return nil, err
	}

	for key, value := range config.GitConfig {
		if _, _, err := repo.Git(append(DefaultGitCommands["config"], key, value)); err != nil {
			return nil, err
		}
	}

	return repo, nil
}
