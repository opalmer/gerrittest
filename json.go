package gerrittest

import "github.com/opalmer/dockertest"

// ServiceSpec is used to serialize information about an instance
// of Gerrit.
type ServiceSpec struct {
	Admin     *User            `json:"admin"`
	Container string           `json:"container"`
	Version   string           `json:"version"`
	SSH       *dockertest.Port `json:"ssh"`
	HTTP      *dockertest.Port `json:"http"`
}
