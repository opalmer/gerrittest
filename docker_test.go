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

func (s *TestSuite) NewClient(c *C, image string) *DockerClient {
	client, err := NewDockerClient(image)
	c.Assert(err, IsNil)
	return client
}

func (s *TestSuite) TestNewContainer_ErrFailedToDeterminePorts(c *C) {
	dockercontainer := types.Container{}
	_, err := NewContainer(dockercontainer)
	c.Assert(err, Equals, ErrPublicPortsMissing)
}

func (s *TestSuite) TestNewDockerClient_DefaultImage(c *C) {
	client := s.NewClient(c, "")
	c.Assert(client.image, Equals, DefaultImage)
}

func (s *TestSuite) TestNewDockerClient_CustomImage(c *C) {
	client := s.NewClient(c, "foo")
	c.Assert(client.image, Equals, "foo")
}

func (s *TestSuite) TestNewDockerClient_BadDockerHost(c *C) {
	defer os.Setenv("DOCKER_HOST", os.Getenv("DOCKER_HOST"))
	os.Setenv("DOCKER_HOST", "none")
	_, err := NewDockerClient("")
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

func (s *TestSuite) TestRunGerrit(c *C) {
	client := s.NewClient(c, "")
	_ = client
	//client.RunGerrit("")

}
