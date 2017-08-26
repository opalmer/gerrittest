package gerrittest

import (
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// Diff is a struct which represents a single commit to the
// repository.
type Diff struct {
	// Error should be set whenever
	Error   error
	Content []byte
}

// ApplyToRoot will apply the given diff to the provided repository root.
func (d *Diff) ApplyToRoot(root string) error {
	if d.Error != nil {
		return d.Error
	}

	cmd := exec.Command(
		"git", "-C", root, "apply")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if _, err := writer.Write(d.Content); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).WithField("out", string(out)).Error()
		return err
	}
	return nil
}
