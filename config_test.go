package gerrittest

import (
	"os"
	"testing"

	"github.com/opalmer/dockertest"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg.Image != DefaultImage {
		t.Fatal()
	}
	if cfg.PortSSH != dockertest.RandomPort {
		t.Fatal()
	}
	if cfg.PortHTTP != dockertest.RandomPort {
		t.Fatal()
	}
	if cfg.CleanupOnFailure != true {
		t.Fatal()
	}

	if value, set := os.LookupEnv("GERRITTEST_DOCKER_IMAGE"); set {
		defer os.Setenv("GERRITTEST_DOCKER_IMAGE", value)
	} else {
		defer os.Unsetenv("GERRITTEST_DOCKER_IMAGE")
	}
	os.Setenv("GERRITTEST_DOCKER_IMAGE", "foo")
	cfgB := NewConfig()
	if cfgB.Image != "foo" {
		t.Fatal()
	}
}
