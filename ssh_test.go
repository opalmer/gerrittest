package gerrittest

import (
	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type SSHTest struct{}

var _ = Suite(&SSHTest{})

func (s *SSHTest) newClient(c *C) (*Service, *SSHClient) {
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)
	cfg := NewConfig()
	cfg.Image = "opalmer/gerrittest:2.14.2"
	service := NewService(client, cfg)
	user, helpers, err := service.Run()
	c.Assert(err, IsNil)
	ssh, err := helpers.GetSSHClient(user)
	c.Assert(err, IsNil)
	return service, ssh
}

func (s *SSHTest) TestVersion(c *C) {
	svc, ssh := s.newClient(c)
	ver, err := ssh.Version()
	c.Assert(err, IsNil)
	c.Assert(ver, Equals, "2.14.2")
	c.Assert(ssh.Close(), IsNil)
	c.Assert(svc.Close(), IsNil)
}

func (s *SSHTest) TestRun(c *C) {
	svc, ssh := s.newClient(c)
	c.Assert(ssh.Close(), IsNil)
	c.Assert(svc.Close(), IsNil)
	//ssh.Run()
}
