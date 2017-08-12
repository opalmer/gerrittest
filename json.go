package gerrittest

import (
	"encoding/json"
	"io/ioutil"

	"github.com/opalmer/dockertest"
	"github.com/spf13/cobra"
)

// ServiceSpec is used to serialize information about an instance
// of Gerrit.
type ServiceSpec struct {
	Admin     *User            `json:"admin"`
	Container string           `json:"container"`
	Version   string           `json:"version"`
	SSH       *dockertest.Port `json:"ssh"`
	HTTP      *dockertest.Port `json:"http"`
}

// ReadServiceSpec reads and returns a *ServiceSpec from a file.
func ReadServiceSpec(path string) (*ServiceSpec, error) {
	var spec ServiceSpec
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// ReadServiceSpecFromArg is identical to ReadServiceSpec except we're
// reading the --json argument from a command.
func ReadServiceSpecFromArg(cmd *cobra.Command) (*ServiceSpec, error) {
	path, err := cmd.Flags().GetString("json")
	if err != nil {
		return nil, err
	}
	return ReadServiceSpec(path)
}
