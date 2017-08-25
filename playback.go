package gerrittest

import (
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	log "github.com/sirupsen/logrus"
)

// DefaultUpstream is the upstream used for pulling in repo repositories.
var DefaultUpstream = "gerrittest"

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
	Read() (<-chan *Diff, error)

	// Cleanup should cleanup any temporary files, directories, etc.
	Cleanup() error
}

// Diff is a struct which represents a single commit to the
// repository.
type Diff struct {
	// Error should be set whenever
	Error error
}

// RemoteRepositorySource reads changes from a remote repository.
type RemoteRepositorySource struct {
	mtx    *sync.Mutex
	log    *log.Entry
	path   string
	remote string
	branch string
}

// Setup will prepare to pull commits from the remote repository.
func (r *RemoteRepositorySource) Setup(*Repository) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	logger := r.log.WithField("phase", "setup")
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	r.path = tempdir

	logger.WithFields(log.Fields{
		"status": "init",
		"path":   tempdir,
	})
	logger.Debug()
	cmd := exec.Command("git", "init", r.path, "--quiet")
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}

	logger = logger.WithFields(log.Fields{
		"status":   "add-remote",
		"path":     tempdir,
		"branch":   r.branch,
		"upstream": r.remote,
	})
	logger.Debug()
	cmd = exec.Command(
		"git", "remote", "add", DefaultUpstream,
		"--track", r.branch, "--mirror=fetch", r.remote)
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}
	// git log --pretty=oneline --format="%H"
	logger = logger.WithField("status", "fetch")
	logger.Debug()
	cmd = exec.Command("git", "fetch", DefaultUpstream, r.branch)
	if data, err := cmd.CombinedOutput(); err != nil {
		logger.WithError(err).WithField("output", string(data)).Error()
		return err
	}

	return nil
}

// Read will read read revisions from the remote repository and play
// them back as diffs into the channel.
func (r *RemoteRepositorySource) Read() (<-chan *Diff, error) {
	diffs := make(chan *Diff, 1)

	// git log FETCH_HEAD --oneline 
	return diffs, nil
}

// Cleanup removes the temporary repository on disk.
func (r *RemoteRepositorySource) Cleanup() error {
	return os.RemoveAll(r.path)
}

// NewRemoteRepositorySource
func NewRemoteRepositorySource(remote string, branch string) (PlaybackSource, error) {
	repo := &RemoteRepositorySource{
		mtx:    &sync.Mutex{},
		log:    log.WithField("cmp", "repo-source"),
		remote: remote,
		branch: branch,
	}
	return repo, nil
}
