package gerrittest

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type ContainerTest struct{}

var _ = Suite(&ContainerTest{})

func (s *ContainerTest) Test_newPort_UnknownPort(c *C) {
	port, err := newPort(dockertest.RandomPort, 5000)
	c.Assert(port, IsNil)
	c.Assert(err, ErrorMatches, "Unknown private port")
}

func (s *ContainerTest) Test_newPort_RandomPortHTTP(c *C) {
	port, err := newPort(dockertest.RandomPort, ExportedHTTPPort)
	c.Assert(err, IsNil)
	c.Assert(port.Public, Not(Equals), dockertest.RandomPort)
	c.Assert(port.Private, Equals, uint16(ExportedHTTPPort))
	c.Assert(port.Protocol, Equals, dockertest.ProtocolTCP)
}

func (s *ContainerTest) Test_newPort_NonRandomPortHTTP(c *C) {
	port, err := newPort(50000, ExportedHTTPPort)
	c.Assert(err, IsNil)
	c.Assert(port.Public, Equals, uint16(50000))
	c.Assert(port.Private, Equals, uint16(ExportedHTTPPort))
	c.Assert(port.Protocol, Equals, dockertest.ProtocolTCP)
}

func (s *ContainerTest) Test_newPort_RandomPortSSH(c *C) {
	port, err := newPort(dockertest.RandomPort, ExportedSSHPort)
	c.Assert(err, IsNil)
	c.Assert(port.Public, Equals, dockertest.RandomPort)
	c.Assert(port.Protocol, Equals, dockertest.ProtocolTCP)
}

func (s *ContainerTest) Test_newPort_NonRandomPortSSH(c *C) {
	port, err := newPort(50000, ExportedSSHPort)
	c.Assert(err, IsNil)
	c.Assert(port.Public, Equals, uint16(50000))
	c.Assert(port.Private, Equals, uint16(ExportedSSHPort))
	c.Assert(port.Protocol, Equals, dockertest.ProtocolTCP)
}

func (s *ContainerTest) TestGetDockerImage_Default(c *C) {
	if value, set := os.LookupEnv(DefaultImageEnvironmentVar); set {
		defer os.Setenv(DefaultImageEnvironmentVar, value) // nolint: errcheck
	}
	c.Assert(os.Unsetenv(DefaultImageEnvironmentVar), IsNil)
	c.Assert(GetDockerImage(""), Equals, DefaultImage)
}

func (s *ContainerTest) TestGetDockerImage_Environment(c *C) {
	if value, set := os.LookupEnv(DefaultImageEnvironmentVar); set {
		defer os.Setenv(DefaultImageEnvironmentVar, value) // nolint: errcheck
		c.Assert(os.Unsetenv(DefaultImageEnvironmentVar), IsNil)
	}
	c.Assert(GetDockerImage(""), Equals, DefaultImage)
}

func (s *ContainerTest) TestGetDockerImage_DirectInput(c *C) {
	c.Assert(GetDockerImage("hello"), Equals, "hello")
}

func (s *ContainerTest) Test_waitPort_contextCancelled(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan error, 1)
	waitPort(ctx, &dockertest.Port{Private: 0, Public: 0, Address: ""}, errs)
	c.Assert(<-errs, ErrorMatches, context.Canceled.Error())
}

func (s *ContainerTest) Test_waitPort(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, IsNil)
	defer listener.Close() // nolint: errcheck
	port, err := strconv.ParseUint(
		strings.Split(listener.Addr().String(), ":")[1], 10, 16)
	c.Assert(err, IsNil)
	errs := make(chan error, 1)
	waitPort(ctx, &dockertest.Port{Private: 0, Public: uint16(port), Address: "127.0.0.1"}, errs)
	c.Assert(<-errs, IsNil)
}

func (s *ContainerTest) Test_waitHTTP_contextCancelled(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan error, 1)
	waitHTTP(ctx, &dockertest.Port{Private: 0, Public: 0, Address: ""}, errs)
	c.Assert(<-errs, ErrorMatches, context.Canceled.Error())
}

func (s *ContainerTest) Test_waitHTTP(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count != 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	port, err := strconv.ParseUint(
		strings.Split(ts.Listener.Addr().String(), ":")[1], 10, 16)
	c.Assert(err, IsNil)

	errs := make(chan error, 1)
	waitHTTP(ctx, &dockertest.Port{Private: 0, Public: uint16(port), Address: "127.0.0.1"}, errs)
	c.Assert(<-errs, IsNil)
}
