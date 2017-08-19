package gerrittest

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/opalmer/dockertest"
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
