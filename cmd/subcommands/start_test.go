package cmd

import (
	"github.com/spf13/cobra"
	. "gopkg.in/check.v1"
)

type StartTest struct{}

var _ = Suite(&StartTest{})

func (s *StartTest) Test_addStartFlags(c *C) {
	command := &cobra.Command{}
	addStartFlags(command)
	c.Assert(command.Flag("password"), NotNil)
}

func (s *StartTest) Test_newStartConfig(c *C) {
	command := &cobra.Command{}
	addStartFlags(command)
	c.Assert(command.ParseFlags(
		[]string{
			"--image=image",
			"--port-http=1",
			"--port-ssh=2",
			"--private-key=private-key",
			"--password=password",
			"--no-cleanup",
			"--start-only",
		}),
		IsNil)
	cfg := newStartConfig(command)
	c.Assert(cfg.Image, Equals, "image")
	c.Assert(cfg.PortHTTP, Equals, uint16(1))
	c.Assert(cfg.PortSSH, Equals, uint16(2))
	c.Assert(cfg.PrivateKeyPath, Equals, "private-key")
	c.Assert(cfg.Password, Equals, "password")
	c.Assert(cfg.Context, NotNil)
	c.Assert(cfg.SkipSetup, Equals, true)
	c.Assert(cfg.SkipCleanup, Equals, true)
}
