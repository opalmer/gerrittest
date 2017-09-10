package gerrittest

import (
	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-gerrit"
	"os"
)

// Change is used to interact with an manipulate a single change.
type Change struct {
	gerrit *Gerrit
	api    *gerrit.Client
	log    *log.Entry
	change     *gerrit.ChangeInfo
}

// ID returns the current change id.
func (c *Change) ID() string {
	return c.change.ID
}

// AddFile adds a file to the repository but does not commit it.
func (c *Change) AddFile(relative string, mode os.FileMode, content []byte) error {
	logger := c.log.WithFields(log.Fields{
		"phase": "add-file",
		"path": relative,
	})
	logger.Debug()
	if err := c.gerrit.Repo.AddContent(relative, mode, content); err != nil {
		logger.WithError(err).Error()
		return err
	}
	c.gerrit.Repo.Commit()

	return nil
}
