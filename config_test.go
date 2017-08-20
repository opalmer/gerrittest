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
	os.Unsetenv(DefaultImageEnvironmentVar)
}

func (s *ConfigTest) TearDownTest(c *C) {
	if s.set {
		os.Setenv(DefaultImageEnvironmentVar, s.value)
		return
	}
	os.Unsetenv(DefaultImageEnvironmentVar)
}

func (s *ConfigTest) TestNewConfigDefaults(c *C) {
	os.Unsetenv(DefaultImageEnvironmentVar)
	cfg := NewConfig()
	c.Assert(cfg, DeepEquals, &Config{
		Image:            DefaultImage,
		PortSSH:          dockertest.RandomPort,
		PortHTTP:         dockertest.RandomPort,
		CleanupOnFailure: true,
	})
}

func (s *ConfigTest) TestNewConfigOverride(c *C) {
	os.Setenv(DefaultImageEnvironmentVar, "override")
	cfg := NewConfig()
	c.Assert(cfg.Image, Equals, "override")
}
