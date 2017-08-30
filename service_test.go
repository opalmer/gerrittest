package gerrittest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
	. "gopkg.in/check.v1"
)

type ServiceTest struct{}

var _ = Suite(&ServiceTest{})

func (s *ServiceTest) TestGetService(c *C) {
	cfg := &Config{
		PortHTTP: dockertest.RandomPort,
	}

	svc, err := GetService(cfg)
	c.Assert(err, IsNil)
	c.Assert(svc.Name, Equals, "gerrittest")
	c.Assert(svc.Timeout, Equals, DefaultStartTimeout)
	c.Assert(cfg.PortHTTP, Not(Equals), dockertest.RandomPort)
	c.Assert(
		svc.Input.Environment, DeepEquals,
		[]string{fmt.Sprintf("GERRIT_CANONICAL_URL=http://127.0.0.1:%d/", cfg.PortHTTP)})
}

func (s *ServiceTest) TestGetRandomPort(c *C) {
	port, err := GetRandomPort()
	c.Assert(err, IsNil)

	// We expect nothing to be listening on the port GetRandomPort()
	// returned. If something is listening then we didn't close the port
	// before leaving GetRandomPort().
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	c.Assert(err, NotNil)
	c.Assert(conn, IsNil)
}

func (s *ServiceTest) TestStart(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}

	svc, err := Start(context.Background(), NewConfig())
	c.Assert(err, IsNil)
	c.Assert(svc.Service.Terminate(), IsNil)
}

func (s *ServiceTest) TestRunner_WaitPortOpen_Cancelled(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{ctx: ctx}
	c.Assert(
		run.waitPortOpen(&dockertest.Port{Address: "127.0.0.1", Public: 0}),
		ErrorMatches, context.Canceled.Error())
}

func (s *ServiceTest) TestRunner_WaitPortOpen_DialError(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	run := runner{ctx: ctx}

	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()

	c.Assert(
		run.waitPortOpen(&dockertest.Port{Address: "127.0.0.1", Public: 65535}),
		ErrorMatches, context.Canceled.Error())
}

func (s *ServiceTest) TestRunner_WaitListenHTTP_Cancelled(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{ctx: ctx}

	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()

	c.Assert(
		run.waitListenHTTP(&dockertest.Port{Address: "127.0.0.1", Public: 65535}),
		ErrorMatches, context.Canceled.Error())
}

func (s *ServiceTest) TestRunner_WaitListenHTTP_BadCode(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{ctx: ctx}
	count := make(chan int, 1)
	count <- 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := <-count
		current++
		count <- current
		if current < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c.Assert(
		run.waitListenHTTP(&dockertest.Port{Address: "127.0.0.1", Public: 65535}),
		ErrorMatches, context.Canceled.Error())
}

func (s *ServiceTest) TestSetup_Err(c *C) {
	setup := Setup{}
	expected := errors.New("testing")
	spec, httpClient, sshClient, err := setup.err(
		log.WithField("phase", "test"), expected)
	c.Assert(spec, IsNil)
	c.Assert(httpClient, IsNil)
	c.Assert(sshClient, IsNil)
	c.Assert(err, ErrorMatches, expected.Error())
}

func (s *ServiceTest) TestSetup_GetKeyPath(c *C) {
	key, err := GenerateRSAKey()
	c.Assert(err, IsNil)
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)

	generatedSigner, err := ssh.NewSignerFromKey(key)
	c.Assert(err, IsNil)
	c.Assert(WriteRSAKey(key, file), IsNil)

	setup := &Setup{PrivateKeyPath: file.Name()}
	_, signer, err := setup.getKey()
	c.Assert(err, IsNil)
	c.Assert(signer.PublicKey().Marshal(), DeepEquals, generatedSigner.PublicKey().Marshal())
	c.Assert(os.Remove(file.Name()), IsNil)
}

// You can start the Gerrit service using the Start() function. This only
// starts the container and returns information about the service.
func ExampleStart() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Terminate the container when you're done.
	if err := svc.Service.Terminate(); err != nil {
		log.Fatal(err)
	}
}

// Once you've started the service you'll want to setup Gerrit inside
// the container. Running Setup.Init will cause the administrative user to
// be created, generate an http api password and insert a public key for ssh
// access.
func ExampleSetup() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	setup := &Setup{Service: svc}
	spec, httpClient, sshClient, err := setup.Init()
	if err != nil {
		log.Fatal(err)
	}
	_ = spec
	_ = httpClient
	_ = sshClient

	// Terminate the container when you're done.
	if err := svc.Service.Terminate(); err != nil {
		log.Fatal(err)
	}
}
