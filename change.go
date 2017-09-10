package gerrittest

import (
	"os"
	"path/filepath"

	"github.com/andygrunwald/go-gerrit"
	log "github.com/sirupsen/logrus"
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
