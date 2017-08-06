package gerrittest

import (
	"errors"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type ServiceTest struct{}

var _ = Suite(&ServiceTest{})

func (s *ServiceTest) TestNewServiceName(c *C) {
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)
	service := NewService(client, NewConfig())
	c.Assert(service.svc.Name, Equals, "gerrittest")
}

func (s *ServiceTest) TestPorts(c *C) {
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)
	service := NewService(client, NewConfig())
	hasSSH := false
	hasHTTP := false
	for _, spec := range service.svc.Input.Ports.Specs {
		if spec.Private == ExportedHTTPPort {
			hasHTTP = true
		}

		if spec.Private == ExportedSSHPort {
			hasSSH = true
		}
	}
	c.Assert(hasSSH, Equals, true)
	c.Assert(hasHTTP, Equals, true)
}

func testRun(client *dockertest.DockerClient, image string, errs chan error) {
	cfg := NewConfig()
	cfg.Image = image
	service := NewService(client, cfg)
	user, helpers, err := service.Run()

	if err != nil {
		service.Close()
		errs <- err
		return
	}

	if user == nil {
		service.Close()
		errs <- errors.New("User not set")
		return
	}

	if helpers == nil {
		service.Close()
		errs <- errors.New("Helpers not set")
		return
	}

	errs <- service.Close()
}

func (s *ServiceTest) TestRun(c *C) {
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)

	images := []string{
		"opalmer/gerrittest:latest",
		"opalmer/gerrittest:2.14.2",
	}
	errs := make(chan error)

	for _, image := range images {
		go testRun(client, image, errs)
	}
	for i := 0; i < len(images); i++ {
		c.Assert(<-errs, IsNil)
	}
}
