package cmd

import (
	"fmt"

	"github.com/opalmer/gerrittest"
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
	generated, err := gerrittest.NewSSHKey()
	c.Assert(err, IsNil)
	defer generated.Remove() // nolint: errcheck
	c.Assert(command.ParseFlags(
		[]string{
			"--image=image",
			"--port-http=1",
			"--port-ssh=2",
			fmt.Sprintf("--private-key=%s", generated.Path),
			"--password=password",
			"--no-cleanup",
			"--start-only",
		}),
		IsNil)
	cfg, err := newStartConfig(command)
	c.Assert(err, IsNil)
	c.Assert(cfg.Image, Equals, "image")
	c.Assert(cfg.PortHTTP, Equals, uint16(1))
	c.Assert(cfg.PortSSH, Equals, uint16(2))

	found := false
	for _, key := range cfg.SSHKeys {
		if key.Path == key.Path {
			found = true
		}
	}

	c.Assert(found, Equals, true)
	c.Assert(cfg.Password, Equals, "password")
	c.Assert(cfg.Context, NotNil)
	c.Assert(cfg.SkipSetup, Equals, true)
}
