package gerrittest

import (
	"os"
	"path/filepath"

	"github.com/andygrunwald/go-gerrit"
	log "github.com/sirupsen/logrus"
)

var (
	// DefaultRevision is the revision to use in the Change struct when
	// no other revision is provided.
	DefaultRevision = "current"
)

// Change is used to interact with an manipulate a single change.
type Change struct {
	gerrit *Gerrit
	api    *gerrit.Client
	log    *log.Entry
	id     string
}

// ID returns the current change id.
func (c *Change) ID() string {
	return c.id
}

// AmendAndPush and push is a small helper that amends the current
// commit then pushes it to Gerrit.
func (c *Change) AmendAndPush() error {
	c.log.WithFields(log.Fields{
		"phase": "amend-and-push",
	}).Debug()
	if err := c.gerrit.Repo.Amend(); err != nil {
		return err
	}
	return c.gerrit.Repo.Push(ProjectName, "")
}

// Write writes a file to the repository but does not commit it. The added or
// modified path will be staged for commit.
func (c *Change) Write(relative string, mode os.FileMode, content []byte) error {
	c.log.WithFields(log.Fields{
		"phase":  "write",
		"path":   relative,
		"length": len(content),
	}).Debug()
	return c.gerrit.Repo.AddContent(relative, mode, content)
}

// Remove will remove the given relative path from the repository. If the file
// or directory does not exist this function does nothing. The removed path
// will staged for commit.
func (c *Change) Remove(relative string) error {
	path := filepath.Join(c.gerrit.Config.RepoRoot, relative)
	c.log.WithFields(log.Fields{
		"phase":    "remove",
		"path":     relative,
		"realpath": path,
	}).Debug()

	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	if stat.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	} else {
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return c.gerrit.Repo.Add(path)
}

// ApplyLabel will apply the requested label to the current change. Examples
// of labels include 'Code-Review +2' or 'Verified +1'. If a specific revision
// is not provided then 'current' will be used.
func (c *Change) ApplyLabel(revision string, label string, value string) (*gerrit.ReviewResult, error) {
	if revision == "" {
		revision = DefaultRevision
	}
	c.log.WithFields(log.Fields{
		"phase":    "apply-label",
		"revision": revision,
		"label":    label,
		"value":    value,
	}).Debug()

	info, _, err := c.api.Changes.SetReview(c.ID(), revision, &gerrit.ReviewInput{
		Labels: map[string]string{
			label: value,
		},
		Drafts: "PUBLISH_ALL_REVISIONS",
	})
	return info, err
}

// AddTopLevelComment will a single top level comment to the current
// change.
func (c *Change) AddTopLevelComment(revision string, comment string) (*gerrit.ReviewResult, error) {
	if revision == "" {
		revision = DefaultRevision
	}
	c.log.WithFields(log.Fields{
		"phase":    "add-top-level-comment",
		"revision": revision,
		"comment":  comment,
	}).Debug()

	result, _, err := c.api.Changes.SetReview(c.ID(), revision, &gerrit.ReviewInput{
		Message:               comment,
		Drafts:                "PUBLISH_ALL_REVISIONS",
		Notify:                "NONE", // Don't send email
		OmitDuplicateComments: true,
	})
	return result, err
}
