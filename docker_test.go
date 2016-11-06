package gerrittest

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	. "gopkg.in/check.v1"
)

type TestSuite struct{}

func TestGerritTest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	TestingT(t)
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) NewClient(c *C) *DockerClient {
	client, err := NewDockerClient()
	c.Assert(err, IsNil)
	return client
}

func (s *TestSuite) TestNewContainer_ErrFailedToDeterminePorts(c *C) {
	dockercontainer := types.Container{}
	_, err := NewContainer(dockercontainer)
	c.Assert(err, Equals, ErrPublicPortsMissing)
}

func (s *TestSuite) TestNewDockerClient_BadDockerHost(c *C) {
	defer os.Setenv("DOCKER_HOST", os.Getenv("DOCKER_HOST"))
	os.Setenv("DOCKER_HOST", "none")
	_, err := NewDockerClient()
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestNewContainer(c *C) {
	dockercontainer := types.Container{
		Ports: []types.Port{
			{PrivatePort: InternalHTTPPort, PublicPort: uint16(50000)},
			{PrivatePort: InternalSSHPort, PublicPort: uint16(60000)}}}
	container, err := NewContainer(dockercontainer)
	c.Assert(err, IsNil)
	c.Assert(container.HTTP, Equals, uint16(50000))
	c.Assert(container.SSH, Equals, uint16(60000))
}

func (s *TestSuite) TestRunAndStopGerrit(c *C) {
	client := s.NewClient(c)
	container, err := client.RunGerrit(nil)
	c.Assert(err, IsNil)
	c.Assert(client.RemoveContainer(container.ID), IsNil)
}

func (s *TestSuite) TestRunWithCustomPort(c *C) {
	client := s.NewClient(c)
	container, err := client.RunGerrit(
		&RunGerritInput{
			PortHTTP: "55500",
			PortSSH:  "44300"})
	defer client.RemoveContainer(container.ID)
	c.Assert(err, IsNil)
	c.Assert(container.HTTP, Equals, uint16(55500))
	c.Assert(container.SSH, Equals, uint16(44300))
}

func (s *TestSuite) TestListContainers(c *C) {
	client := s.NewClient(c)
	container, err := client.RunGerrit(nil)
	defer client.RemoveContainer(container.ID)
	c.Assert(err, IsNil)

	containers, err := client.Containers()
	c.Assert(err, IsNil)

	found := false
	for _, listed := range containers {
		if listed.ID == container.ID {
			found = true
			break
		}
	}
	c.Assert(found, Equals, true)

}

func (s *TestSuite) TestGetContainer(c *C) {
	client := s.NewClient(c)
	container, err := client.RunGerrit(nil)
	defer client.RemoveContainer(container.ID)
	c.Assert(err, IsNil)

	containers, err := client.Containers()
	c.Assert(err, IsNil)

	found := false
	for _, listed := range containers {
		if listed.ID == container.ID {
			found = true
			break
		}
	}
	c.Assert(found, Equals, true)
}
