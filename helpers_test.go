package gerrittest

import (
	"os"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type HelpersTest struct{}

var _ = Suite(&HelpersTest{})

func (s *HelpersTest) newHelpers(c *C) (*Helpers, *Service, *User) {
	client, err := dockertest.NewClient()
	c.Assert(err, IsNil)
	cfg := NewConfig()
	cfg.Image = "opalmer/gerrittest:2.14.2"
	service := NewService(client, cfg)
	user, helpers, err := service.Run()
	c.Assert(err, IsNil)
	return helpers, service, user
}

func (s *HelpersTest) TestGetURL(c *C) {
	helpers := NewHelpers(&dockertest.Port{
		Address: "0.0.0.0", Public: 5432,
	}, nil)
	c.Assert(
		helpers.GetURL("/foo/bar"), Equals,
		"http://0.0.0.0:5432/foo/bar")
}

func (s *HelpersTest) TestCreateSSHKeyPair(c *C) {
	helpers := NewHelpers(nil, nil)
	publicPath, privatePath, err := helpers.CreateSSHKeyPair()
	c.Assert(err, IsNil)
	_, err = os.Stat(publicPath)
	c.Assert(err, IsNil)
	_, err = os.Stat(privatePath)
	c.Assert(err, IsNil)
}

func (s *HelpersTest) TestGetSSHClient(c *C) {
	helpers, svc, user := s.newHelpers(c)
	client, err := helpers.GetSSHClient(user)
	c.Assert(err, IsNil)
	c.Assert(client.Close(), IsNil)
	c.Assert(svc.Close(), IsNil)
}

func (s *HelpersTest) TestCheckHTTPLogin(c *C) {
	helpers, svc, user := s.newHelpers(c)
	c.Assert(
		helpers.CheckHTTPLogin(user.Login, user.Password), IsNil)
	c.Assert(svc.Close(), IsNil)
}
