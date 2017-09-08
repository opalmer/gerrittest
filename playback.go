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

	log "github.com/Sirupsen/logrus"
)

// DefaultUpstream is the upstream used for pulling in repo repositories.
var DefaultUpstream = "upstream-gerrittest"

// PlaybackSource is an interface which when implemented can be used
// to playback changes from some kind of source. This intended to
// be used in conjunction with *Repository.
type PlaybackSource interface {
	// Setup is responsible for configuring the playback source prior to Read()
	Setup(repo *Repository) error

	// Read should read the history from a source source and produce
	// a channel containing diffs as well as an error channel. The diff
	// channel should be closed when when there are no further diffs to
	// be played back.
	Read(ctx context.Context) (<-chan *Diff, error)

	// Cleanup should cleanup any temporary files, directories, etc.
	Cleanup() error
}

// RemoteRepositorySource reads changes from a remote repository.
type RemoteRepositorySource struct {
	log    *log.Entry
	path   string
	remote string
	branch string
}

// Setup does nothing for this source.
func (r *RemoteRepositorySource) Setup(repo *Repository) error {
	return nil
}

// Read will read read revisions from the remote repository and play
// them back as diffs into the channel.
func (r *RemoteRepositorySource) Read(ctx context.Context) (<-chan *Diff, error) {
	logger := r.log.WithField("phase", "read")
	diffs := make(chan *Diff, 1)

	// First, get a list of all commits in the remote.
	cmd := exec.CommandContext(
		ctx, "git", "-C", r.path, "log", "FETCH_HEAD", "--oneline")
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

		// Play the commits back in reverse.
		for i := len(commits) - 1; i >= 0; i-- {
			commit := commits[i]
			entry.WithField("commit", commit).Debug()
			cmd := exec.CommandContext(
				ctx, "git", "-C", r.path, "format-patch", "-1", commit,
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
			diffs <- &Diff{
				Commit:  commit,
				Content: data,
			}
			count++
		}
	}()

	return diffs, nil
}

// Cleanup removes the temporary repository on disk.
func (r *RemoteRepositorySource) Cleanup() error {
	return os.RemoveAll(r.path)
}

// NewRemoteRepositorySource constructs and returns a *RemoteRepositorySource
// which is an implementation of PlaybackSource.
func NewRemoteRepositorySource(remote string, branch string) (PlaybackSource, error) {
	logger := log.WithField("cmp", "playback")
	tempdir, err := ioutil.TempDir("", fmt.Sprintf("%s-playback-", ProjectName))
	if err != nil {
		return nil, err
	}

	entry := logger.WithField("phase", "setup")
	entry.WithFields(log.Fields{
		"status": "init",
		"path":   tempdir,
	})
	entry.Debug()

	cmd := exec.Command("git", "init", tempdir, "--quiet")
	if data, err := cmd.CombinedOutput(); err != nil {
		entry.WithError(err).WithField("output", string(data)).Error()
		return nil, err
	}

	entry = entry.WithFields(log.Fields{
		"status":   "add-remote",
		"path":     tempdir,
		"branch":   branch,
		"upstream": remote,
	})
	entry.Debug()

	cmd = exec.Command(
		"git", "-C", tempdir, "remote", "add", DefaultUpstream,
		"--track", fmt.Sprintf("refs/heads/%s", branch),
		"--mirror=fetch", remote)
	if data, err := cmd.CombinedOutput(); err != nil {
		entry.WithError(err).WithField("output", string(data)).Error()
		return nil, err
	}

	entry = logger.WithField("status", "fetch")
	entry.Debug()
	cmd = exec.Command(
		"git", "-C", tempdir, "fetch", DefaultUpstream, branch)
	if data, err := cmd.CombinedOutput(); err != nil {
		entry.WithError(err).WithField("output", string(data)).Error()
		return nil, err
	}
	repo := &RemoteRepositorySource{
		log:    logger,
		path:   tempdir,
		remote: remote,
		branch: branch,
	}
	return repo, nil
}
