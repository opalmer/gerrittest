package gerrittest

import (
	. "gopkg.in/check.v1"
	"os"
	"os/exec"
)

func (s *TestSuite) TestGetAddress_Default(c *C) {
	if value, set := os.LookupEnv("DOCKER_HOST"); set {
		defer os.Setenv("DOCKER_HOST", value)
	}

	command := exec.Command("docker-machine", "ip")
	expected := "127.0.0.1"
	if data, err := command.Output(); err == nil {
		expected = AddressOnly(string(data))
	}
	os.Unsetenv("DOCKER_HOST")
	c.Assert(GetAddress(), Equals, expected)
}

func (s *TestSuite) TestGetAddress_DockerHost(c *C) {
	originalValue, set := os.LookupEnv("DOCKER_HOST")
	if set {
		defer os.Setenv("DOCKER_HOST", originalValue)
	} else {
		defer os.Unsetenv("DOCKER_HOST")
	}

	os.Setenv("DOCKER_HOST", "foo")
	c.Assert(GetAddress(), Equals, "foo")
}

func (s *TestSuite) TestAddressOnly(c *C) {
	results := map[string]string{
		"127.0.0.1": "127.0.0.1",
		"foo":       "foo",
		"tcp://127.0.0.1:4242/": "127.0.0.1",
	}

	for input, output := range results {
		c.Assert(AddressOnly(input), Equals, output)
	}
}
