package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/opalmer/gerrittest"
	. "gopkg.in/check.v1"
)

type StopTest struct{}

var _ = Suite(&StopTest{})

func (s *StopTest) TestStop(c *C) {
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
	c.Assert(Stop.Flags().Parse([]string{"--json", file.Name()}), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), IsNil)
}

func (s *StopTest) TestStopBadSpec(c *C) {
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	c.Assert(Stop.Flags().Parse([]string{"--json", file.Name()}), IsNil)
	_, err = file.WriteString("{")
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), ErrorMatches, "unexpected end of JSON input")
}

func (s *StopTest) TestStopJSONFlagNotProvided(c *C) {
	c.Assert(Stop.Flags().Parse([]string{}), IsNil)
	Stop.Flags().Set("json", "")
	c.Assert(Stop.RunE(Stop, []string{}), ErrorMatches, "--json not provided")
}
