package gerrittest

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/andygrunwald/go-gerrit"
	log "github.com/sirupsen/logrus"
)

var (
	// DefaultRevision is the revision to use in the Change struct when
	// no other revision is provided.
	DefaultRevision = "current"
)

const (
	// VerifiedLabel is a string representing the 'Verified' label
	VerifiedLabel = "Verified"

	// CodeReviewLabel is a string representing the 'Code-Review' label
	CodeReviewLabel = "Code-Review"
)

// Change is used to interact with an manipulate a single change.
type Change struct {
	api      *gerrit.Client
	log      *log.Entry
	ChangeID string
	Repo     *Repository
}

func (c *Change) logError(err error, logger *log.Entry, response *gerrit.Response) {
	if err != nil {
		logger = logger.WithError(err)
		body, _ := ioutil.ReadAll(response.Body) // nolint: errcheck
		logger.WithField("body", string(body)).Error()
	}
}

// Destroy is responsible for removing any files on disk related to
// this change.
func (c *Change) Destroy() error {
	return c.Repo.Destroy()
}

// Push pushes changes to Gerrit.
func (c *Change) Push() error {
	return c.Repo.Push("HEAD:refs/for/master")
}

// Add writes a file to the repository but does not commit it. The added or
// modified path will be staged for commit.
func (c *Change) Add(relative string, mode os.FileMode, content string) error {
	if err := c.Repo.Add(relative, mode, []byte(content)); err != nil {
		return err
	}
	return c.Repo.Amend()
}

// Remove will remove the given relative path from the repository. If the file
// or directory does not exist this function does nothing. The removed path
// will staged for commit.
func (c *Change) Remove(relative string) error {
	if err := c.Repo.Remove(relative); err != nil {
		return err
	}
	return c.Repo.Amend()
}

// ApplyLabel will apply the requested label to the current change. Examples
// of labels include 'Code-Review +2' or 'Verified +1'. If a specific revision
// is not provided then 'current' will be used.
func (c *Change) ApplyLabel(revision string, label string, value int) (*gerrit.ReviewResult, error) {
	if revision == "" {
		revision = DefaultRevision
	}
	logger := c.log.WithFields(log.Fields{
		"phase":    "apply-label",
		"revision": revision,
		"label":    label,
		"value":    value,
	})

	logger = logger.WithField("id", c.ChangeID)
	logger.Debug()

	info, response, err := c.api.Changes.SetReview(c.ChangeID, revision, &gerrit.ReviewInput{
		Labels: map[string]string{
			label: strconv.Itoa(value),
		},
		Drafts: "PUBLISH_ALL_REVISIONS",
	})
	c.logError(err, logger, response)
	return info, err
}

// Submit will submit the change. Note, this typically will only work if the
// change has Code-Review +2 and Verified +1 labels applied.
func (c *Change) Submit() (*gerrit.ChangeInfo, error) {
	logger := c.log.WithField("phase", "submit")
	logger.Debug()
	info, response, err := c.api.Changes.SubmitChange(c.ChangeID, &gerrit.SubmitInput{})
	c.logError(err, logger, response)
	return info, err
}

// Abandon will abandon the change.
func (c *Change) Abandon() (*gerrit.ChangeInfo, error) {
	logger := c.log.WithField("phase", "abandon")
	logger.Debug()
	info, response, err := c.api.Changes.AbandonChange(c.ChangeID, &gerrit.AbandonInput{
		Notify: "NONE",
	})
	c.logError(err, logger, response)
	return info, err
}

// AddTopLevelComment will a single top level comment to the current
// change.
func (c *Change) AddTopLevelComment(revision string, comment string) (*gerrit.ReviewResult, error) {
	if revision == "" {
		revision = DefaultRevision
	}
	logger := c.log.WithFields(log.Fields{
		"phase":    "add-top-level-comment",
		"revision": revision,
		"comment":  comment,
	})
	id, err := c.Repo.ChangeID()
	if err != nil {
		return nil, err
	}
	logger = logger.WithField("id", id)
	logger.Debug()

	result, response, err := c.api.Changes.SetReview(id, revision, &gerrit.ReviewInput{
		Message:               comment,
		Drafts:                "PUBLISH_ALL_REVISIONS",
		Notify:                "NONE", // Don't send email
		OmitDuplicateComments: true,
	})
	c.logError(err, logger, response)
	return result, err
}

// AddFileComment will apply a comment to a specific file in a specific
// location
func (c *Change) AddFileComment(revision string, path string, line int, comment string) (*gerrit.ReviewResult, error) {
	if revision == "" {
		revision = DefaultRevision
	}
	logger := c.log.WithFields(log.Fields{
		"phase":    "add-file-comment",
		"revision": revision,
		"comment":  comment,
	})
	comments := map[string][]gerrit.CommentInput{}
	comments[path] = append(comments[path], gerrit.CommentInput{
		Message: comment,
		Line:    line,
		Side:    "REVISION",
		Range: gerrit.CommentRange{
			StartLine: line,
			EndLine:   line,
		},
	})
	result, response, err := c.api.Changes.SetReview(c.ChangeID, revision, &gerrit.ReviewInput{
		Comments:              comments,
		Drafts:                "PUBLISH_ALL_REVISIONS",
		Notify:                "NONE", // Don't send email
		OmitDuplicateComments: true,
	})
	c.logError(err, logger, response)
	return result, err
}
