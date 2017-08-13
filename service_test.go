package gerrittest

import (
	"time"

	"fmt"
	"net"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type RunnerTest struct{}

var _ = Suite(&RunnerTest{})

func (s *RunnerTest) TestGetService(c *C) {
	cfg := &Config{
		PortHTTP: dockertest.RandomPort,
	}
	svc, err := GetService(cfg)
	c.Assert(len(svc.Input.Ports.Specs), Equals, 2)
	c.Assert(err, IsNil)
	c.Assert(svc.Name, Equals, "gerrittest")
	c.Assert(svc.Timeout, DeepEquals, time.Minute*10)
	if cfg.PortHTTP == dockertest.RandomPort {
		c.Fail()
	}
	c.Assert(len(svc.Input.Environment), Equals, 1)
}

func (s *RunnerTest) TestGetRandomPort(c *C) {
	port, err := GetRandomPort()
	c.Assert(err, IsNil)
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	c.Assert(err, NotNil)
	c.Assert(conn, IsNil)
}
