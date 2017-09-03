package cmd

import (
	"io/ioutil"

	. "gopkg.in/check.v1"
)

type StopTest struct{}

var _ = Suite(&StopTest{})

func (s *StopTest) TestStop_BadSpec(c *C) {
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	c.Assert(Stop.Flags().Parse([]string{"--json", file.Name()}), IsNil)
	_, err = file.WriteString("{")
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), ErrorMatches, "unexpected end of JSON input")
}

func (s *StopTest) TestStop_JSONFlagNotProvided(c *C) {
	c.Assert(Stop.Flags().Parse([]string{}), IsNil)
	c.Assert(Stop.Flags().Set("json", ""), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), ErrorMatches, "--json not provided")
}
