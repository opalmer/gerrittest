package gerrittest

import (
	"io/ioutil"
	"os"

	"github.com/go-ini/ini"
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

func (s *ConfigTest) TestNewConfigOverride(c *C) {
	c.Assert(os.Setenv(DefaultImageEnvironmentVar, "override"), IsNil)
	cfg := NewConfig()
	c.Assert(cfg.Image, Equals, "override")
}

func (s *ConfigTest) testWrittenConfig(c *C, path string) {
	cfg, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, path)
	c.Assert(err, IsNil)
	heads := cfg.Section(accessHeads)
	c.Assert(
		heads.Key("label-Verified").ValueWithShadows(), DeepEquals,
		[]string{"-1..+1 group Administrators", "-1..+1 group Project Owners"})
	verifiedLabel := cfg.Section(labelVerified)
	c.Assert(verifiedLabel.Key("function").Value(), Equals, "MaxWithBlock")
	c.Assert(verifiedLabel.Key("defaultValue").Value(), Equals, "0")
	c.Assert(
		verifiedLabel.Key("value").ValueWithShadows(), DeepEquals,
		[]string{"-1 Fails", "0 No Score", "+1 Verified"})
}

func (s *ConfigTest) Test_projectConfig_newProjectConfig_missingFile(c *C) {
	file, err := ioutil.TempFile("", "")
	defer os.Remove(file.Name()) // nolint: errcheck
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	c.Assert(os.Remove(file.Name()), IsNil)
	cfg, err := newProjectConfig(file.Name())
	c.Assert(err, IsNil)
	c.Assert(cfg.write(file.Name()), IsNil)
	s.testWrittenConfig(c, file.Name())
}

func (s *ConfigTest) Test_projectConfig_newProjectConfig_existingConifg(c *C) {
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	_, err = file.WriteString(`
[label "Verified"]
default = 1
	`)
	c.Assert(err, IsNil)
	defer os.Remove(file.Name()) // nolint: errcheck
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	cfg, err := newProjectConfig(file.Name())
	c.Assert(err, IsNil)
	c.Assert(cfg.write(file.Name()), IsNil)
	s.testWrittenConfig(c, file.Name())
}
