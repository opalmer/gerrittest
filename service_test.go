package gerrittest

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/opalmer/dockertest"
	"github.com/prometheus/common/log"
)

func TestGetService(t *testing.T) {
	cfg := &Config{
		PortHTTP: dockertest.RandomPort,
	}

	svc, err := GetService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if svc.Name != "gerrittest" {
		t.Fatal()
	}
	if svc.Timeout != time.Minute*10 {
		t.Fatal()
	}
	if cfg.PortHTTP == dockertest.RandomPort {
		t.Fatal()
	}
	if len(svc.Input.Environment) != 1 {
		t.Fatal()
	}
}

func TestGetRandomPort(t *testing.T) {
	port, err := GetRandomPort()
	if err != nil {
		t.Fatal(err)
	}

	// We expect nothing to be listening on the port GetRandomPort()
	// returned. If something is listening then we didn't close the port
	// before leaving GetRandomPort().
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		t.Fatal()
	}
	if conn != nil {
		t.Fatal()
	}
}

func TestStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Service.Terminate()
}

// You can start the Gerrit service using the Start() function. This only
// starts the container and returns information about the service.
func ExampleStart() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Service.Terminate() // Terminate the container when you're done.
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
	defer svc.Service.Terminate() // Terminate the container when you're done.

	setup := &Setup{Service: svc}
	spec, http, ssh, err := setup.Init()
	if err != nil {
		log.Fatal(err)
	}
	_ = spec
	_ = http
	_ = ssh
}
