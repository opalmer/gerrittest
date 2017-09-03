package gerrittest

import (
	"os"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type ConfigTest struct {
	value string
	set   bool
}

var _ = Suite(&ConfigTest{})

func (s *ConfigTest) SetUpTest(c *C) {
	value, set := os.LookupEnv(DefaultImageEnvironmentVar)
	s.value = value
	s.set = set
	c.Assert(os.Unsetenv(DefaultImageEnvironmentVar), IsNil)
}

func (s *ConfigTest) TearDownTest(c *C) {
	if s.set {
		c.Assert(os.Setenv(DefaultImageEnvironmentVar, s.value), IsNil)
		return
	}
	c.Assert(os.Unsetenv(DefaultImageEnvironmentVar), IsNil)
}

func (s *ConfigTest) TestNewConfigDefaults(c *C) {
	c.Assert(os.Unsetenv(DefaultImageEnvironmentVar), IsNil)
	cfg := NewConfig()
	c.Assert(cfg, DeepEquals, &Config{
		Image:    DefaultImage,
		PortSSH:  dockertest.RandomPort,
		PortHTTP: dockertest.RandomPort,
		RepoRoot: "",
		Username: "admin",
		Password: "",
	})
}

func (s *ConfigTest) TestNewConfigOverride(c *C) {
	c.Assert(os.Setenv(DefaultImageEnvironmentVar, "override"), IsNil)
	cfg := NewConfig()
	c.Assert(cfg.Image, Equals, "override")
}
