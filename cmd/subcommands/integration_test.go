package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/opalmer/gerrittest"
	. "gopkg.in/check.v1"
)

type IntegrationTest struct{}

var _ = Suite(&IntegrationTest{})

func (s *IntegrationTest) Test_StartStop(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	defer os.Remove(file.Name()) // nolint: errcheck
	jsonFlag := []string{fmt.Sprintf("--json=%s", file.Name())}

	// Call the start subcommand and write the struct to disk.
	c.Assert(file.Close(), IsNil)
	c.Assert(Start.ParseFlags(jsonFlag), IsNil)
	c.Assert(Start.RunE(Start, []string{}), IsNil)

	// Loading
	gerrit, err := gerrittest.NewFromJSON(file.Name())
	c.Assert(err, IsNil)
	c.Assert(gerrit.HTTP.CreateProject("testing"), IsNil)

	// Read the struct from disk and use it to stop the container
	c.Assert(Stop.ParseFlags(jsonFlag), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), IsNil)
}
