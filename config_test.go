package gerrittest

import (
	"testing"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type ConfigTest struct{}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&ConfigTest{})

func (s *ConfigTest) TestNewConfig(c *C) {
	c.Assert(NewConfig(), DeepEquals, &Config{
		Image:       "opalmer/gerrittest:latest",
		PortSSH:     dockertest.RandomPort,
		PortHTTP:    dockertest.RandomPort,
		CreateAdmin: true,
	})
}
