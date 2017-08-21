package cmd

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	. "gopkg.in/check.v1"
)

type UtilTest struct{}

var _ = Suite(&UtilTest{})

func (s *UtilTest) SetUpTest(c *C) {
	exit = false
}

func (s *UtilTest) TearDownTest(c *C) {
	exit = true
}

func (s *UtilTest) TestExitIf(c *C) {
	c.Assert(exitIf("", nil), Equals, false)
	c.Assert(exitIf("", errors.New("hello")), Equals, true)
}

func (s *UtilTest) TestGetBool(c *C) {
	command := &cobra.Command{}
	command.Flags().Bool("test", false, "")
	command.ParseFlags([]string{"--test"})
	c.Assert(getBool(command, "test"), Equals, true)
}

func (s *UtilTest) TestGetString(c *C) {
	command := &cobra.Command{}
	command.Flags().String("test", "", "")
	command.ParseFlags([]string{"--test=bar"})
	c.Assert(getString(command, "test"), Equals, "bar")
}

func (s *UtilTest) TestGetDuration(c *C) {
	command := &cobra.Command{}
	command.Flags().Duration("test", time.Second*0, "")
	command.ParseFlags([]string{"--test=15m"})
	c.Assert(getDuration(command, "test"), Equals, time.Minute*15)
}

func (s *UtilTest) TestGetUInt16(c *C) {
	command := &cobra.Command{}
	command.Flags().Uint16("test", 0, "")
	command.ParseFlags([]string{"--test=65535"})
	c.Assert(getUInt16(command, "test"), Equals, uint16(65535))
}