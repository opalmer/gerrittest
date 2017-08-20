package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	. "gopkg.in/check.v1"
)

type StartTest struct{}

var _ = Suite(&StartTest{})

func (s *StartTest) TestStart(c *C) {
	if testing.Short() {
		c.Skip("-skip set")
	}
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	c.Assert(os.Remove(file.Name()), IsNil)
	c.Assert(Start.Flags().Parse([]string{"--json", file.Name()}), IsNil)
	c.Assert(Start.RunE(Start, []string{}), IsNil)
	output, err := ioutil.ReadFile(file.Name())
	c.Assert(err, IsNil)
	spec := &gerrittest.ServiceSpec{}
	c.Assert(json.Unmarshal(output, spec), IsNil)
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)
	c.Assert(client.RemoveContainer(context.Background(), spec.Container), IsNil)
}
