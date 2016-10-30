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

func (s *TestSuite) TestNewContainer_ErrFailedToDeterminePorts(c *C) {
	dockercontainer := types.Container{}
	_, err := NewContainer(dockercontainer)
	c.Assert(err, Equals, ErrPublicPortsMissing)
}

func (s *TestSuite) TestNewDockerClient(c *C) {
	_, err := NewDockerClient()
	c.Assert(err, IsNil)
}

func (s *TestSuite) TestNewDockerClient_Err(c *C) {
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
