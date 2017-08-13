package gerrittest

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type HTTPTest struct{}

var _ = Suite(&HTTPTest{})

func newService(c *C) *Service {
	httpPort, err := GetRandomPort()
	c.Assert(err, IsNil)
	sshPort, err := GetRandomPort()
	c.Assert(err, IsNil)

	return &Service{
		HTTPPort: &dockertest.Port{
			Address:  "127.0.0.1",
			Public:   httpPort,
			Private:  ExportedHTTPPort,
			Protocol: dockertest.ProtocolTCP,
		},
		SSHPort: &dockertest.Port{
			Address:  "127.0.0.1",
			Public:   sshPort,
			Private:  ExportedSSHPort,
			Protocol: dockertest.ProtocolTCP,
		},
	}
}

func (s *HTTPTest) TestNewHTTPClient(c *C) {
	svc := newService(c)
	client, err := NewHTTPClient(svc, "foobar")
	c.Assert(err, IsNil)
	expected := &HTTPClient{
		Client: &http.Client{Jar: NewCookieJar()},
		Prefix: fmt.Sprintf(
			"http://%s:%d", svc.HTTPPort.Address, svc.HTTPPort.Public),
		User: "foobar",
		log:  log.WithField("cmp", "http"),
	}
	c.Assert(client, DeepEquals, expected)
}
