package gerrittest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// DefaultUpstream is the upstream used for pulling in repo repositories.
var DefaultUpstream = "upstream-gerrittest"

// PlaybackSource is an interface which when implemented can be used
// to playback changes from some kind of source. This intended to
// be used in conjunction with *Repository.
type PlaybackSource interface {
	// Setup is used to prepare the destination repository and/or
	// setup the playback.
	Setup(*Repository) error

	// Read should read the history from a source source and produce
	// a channel containing diffs as well as an error channel. The diff
	// channel should be closed when when there are no further diffs to
	// be played back.
	Read(ctx context.Context) (<-chan *Diff, error)

	// Cleanup should cleanup any temporary files, directories, etc.
	Cleanup() error
}

// Diff is a struct which represents a single commit to the
// repository.
type Diff struct {
	// Error should be set whenever
	Error   error
	Content []byte
}

// RemoteRepositorySource reads changes from a remote repository.
type RemoteRepositorySource struct {
	log    *log.Entry
	repo   string
	remote string
	branch string
}

// Setup will prepare to pull commits from the remote repository.
func (r *RemoteRepositorySource) Setup(*Repository) error {
	logger := r.log.WithField("phase", "setup")
	logger.WithFields(log.Fields{
		"status": "init",
		"path":   r.repo,
	})
	logger.Debug()
	cmd := exec.Command("git", "init", r.repo, "--quiet")
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}

	logger = logger.WithFields(log.Fields{
		"status":   "add-remote",
		"path":     r.repo,
		"branch":   r.branch,
		"upstream": r.remote,
	})
	logger.Debug()

	cmd = exec.Command(
		"git", "-C", r.repo, "remote", "add", DefaultUpstream,
		"--track", fmt.Sprintf("refs/heads/%s", r.branch),
		"--mirror=fetch", r.remote)
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}

	logger = logger.WithField("status", "fetch")
	logger.Debug()
	cmd = exec.Command(
		"git", "-C", r.repo, "fetch", DefaultUpstream, r.branch)
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}

	return nil
}

// Read will read read revisions from the remote repository and play
// them back as diffs into the channel.
func (r *RemoteRepositorySource) Read(ctx context.Context) (<-chan *Diff, error) {
	logger := r.log.WithField("phase", "read")
	diffs := make(chan *Diff, 1)

	// First, get a list of all commits in the remote.
	cmd := exec.Command(
		"git", "-C", r.repo, "log", "FETCH_HEAD", "--oneline")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	entry := logger.WithField("status", "read-commits")
	entry.Debug()
	commitScanner := bufio.NewScanner(bytes.NewReader(output))
	commits := []string{}
	for commitScanner.Scan() {
		commit := strings.Split(commitScanner.Text(), " ")[0]
		commits = append(commits, commit)
	}
	entry.WithFields(log.Fields{
		"commits":  len(commits),
		"duration": time.Since(start),
	}).Debug()

	go func() {
		entry := logger.WithField("status", "extract")
		start := time.Now()
		count := 0
		defer func() {
			close(diffs)
			entry.WithFields(log.Fields{
				"progress": fmt.Sprintf("%d/%d", count, len(commits)),
				"duration": time.Since(start),
			}).Debug()
		}()

		for _, commit := range commits {
			entry.WithField("commit", commit).Debug()
			cmd := exec.Command(
				"git", "-C", r.repo, "format-patch", "-1", commit,
				"--stdout")
			select {
			case <-ctx.Done():
				return
			default:
			}

			data, err := cmd.CombinedOutput()
			if err != nil {
				diffs <- &Diff{Error: err}
				continue
			}
			diffs <- &Diff{Content: data}
			count++
		}
	}()

	return diffs, nil
}

// Cleanup removes the temporary repository on disk.
func (r *RemoteRepositorySource) Cleanup() error {
	return os.RemoveAll(r.repo)
}

// NewRemoteRepositorySource
func NewRemoteRepositorySource(remote string, branch string) (PlaybackSource, error) {
	logger := log.WithField("cmp", "repo-source")
	tempdir, err := ioutil.TempDir("", "gerrittest-")
	if err != nil {
		return nil, err
	}
	repo := &RemoteRepositorySource{
		log:    logger,
		repo:   tempdir,
		remote: remote,
		branch: branch,
	}
	return repo, nil
}
