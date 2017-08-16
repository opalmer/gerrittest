package gerrittest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
)

const (
	// ExportedHTTPPort is the port exported by the docker container
	// where the HTTP service is running.
	ExportedHTTPPort = 8080

	// ExportedSSHPort is the port exported by the docker container
	// where the SSHPort service is running.
	ExportedSSHPort = 29418
)

// User represents a single user for connecting to Gerrit.
type User struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

// Service is a struct representing a running instance of Gerrit
// inside of a docker container.
type Service struct {
	Container *dockertest.ContainerInfo
	HTTPPort  *dockertest.Port
	SSHPort   *dockertest.Port
}

// HTTPClient constructs and returns a basic client for interacting with
// the service.
func (s *Service) HTTPClient() (*HTTPClient, error) {
	return NewHTTPClient(s, "admin")
}

// runner used to run Gerrit and wait for it to come up.
type runner struct {
	ctx    context.Context
	cfg    *Config
	cancel context.CancelFunc
	log    *log.Entry
	http   *dockertest.Port
	ssh    *dockertest.Port
}

func (s *runner) waitPortOpen(port *dockertest.Port) error {
	addr := fmt.Sprintf("%s:%d", port.Address, port.Public)
	ticker := time.NewTicker(time.Millisecond * 200)
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case <-ticker.C:
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				conn.Close()
				continue
			}
			return conn.Close()
		}
	}
}

func (s *runner) waitListenHTTP(port *dockertest.Port) error {
	url := fmt.Sprintf("http://%s:%d", port.Address, port.Public)
	ticker := time.NewTicker(time.Millisecond * 200)
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case <-ticker.C:
			response, err := http.Get(url + "/")
			if err != nil {
				continue
			}
			if response.StatusCode != http.StatusOK {
				continue
			}
			return nil
		}
	}
}

// ping waits for the container to come up. Any error returned from this
// function will cause dockertest to to terminate the underlying container.
func (s *runner) ping(input *dockertest.PingInput) error {
	portSSH, err := input.Container.Port(ExportedSSHPort)
	if err != nil {
		return err
	}
	portHTTP, err := input.Container.Port(ExportedHTTPPort)
	if err != nil {
		return err
	}

	// Wait for both sockets to be listening
	start := time.Now()
	logger := s.log.WithFields(log.Fields{
		"phase": "wait-tcp-sockets",
	})

	logger.Debug()
	errs := make(chan error, 2)
	go func() { errs <- s.waitPortOpen(portSSH) }()
	go func() { errs <- s.waitPortOpen(portHTTP) }()
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			return err
		}
	}
	logger.WithField("duration", time.Since(start)).Debug()

	// Wait for GET / to return 200
	start = time.Now()
	logger = s.log.WithFields(log.Fields{
		"phase": "wait-http",
	})
	err = s.waitListenHTTP(portHTTP)
	logger.WithField("duration", time.Since(start)).Debug()
	if err != nil {
		logger.WithError(err).Error()
	}
	s.http = portHTTP
	s.ssh = portSSH
	return err
}

// GetRandomPort will make its best effort to pick a random port
// to bind to.
func GetRandomPort() (uint16, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	port, err := strconv.ParseUint(
		strings.Split(listener.Addr().String(), ":")[1], 10, 16)
	return uint16(port), err
}

// GetService takes a given configuration and returns the resulting
// dockertest.Service struct.
func GetService(cfg *Config) (*dockertest.Service, error) {
	dockerService, err := dockertest.NewClient()
	if err != nil {
		return nil, err
	}

	// If a random port has been chosen then we need to try and determine
	// one instead of letting Docker choose. If we don't, we'll be unable
	// to properly set $GERRIT_CANONICAL_URL down below.
	if cfg.PortHTTP == dockertest.RandomPort {
		port, err := GetRandomPort()
		if err != nil {
			return nil, err
		}
		cfg.PortHTTP = port
	}

	input := dockertest.NewClientInput(cfg.Image)
	input.Ports.Add(&dockertest.Port{
		Private:  ExportedHTTPPort,
		Public:   cfg.PortHTTP,
		Protocol: dockertest.ProtocolTCP,
	})
	input.Ports.Add(&dockertest.Port{
		Private:  ExportedSSHPort,
		Public:   cfg.PortSSH,
		Protocol: dockertest.ProtocolTCP,
	})
	svc := dockerService.Service(input)
	svc.Name = "gerrittest"
	svc.Timeout = time.Minute * 10 // handled externally

	svc.Input.AddEnvironmentVar(
		"GERRIT_CANONICAL_URL",
		fmt.Sprintf("http://127.0.0.1:%d/", cfg.PortHTTP))
	return svc, nil
}

// Start will start Gerrit and return a struct containing information
// on how to connect.
func Start(parent context.Context, cfg *Config) (*Service, error) {
	service, err := GetService(cfg)
	if err != nil {
		return nil, err
	}

	logger := log.WithField("svc", "runner")
	ctx, cancel := context.WithCancel(parent)
	run := &runner{
		ctx:    ctx,
		log:    logger,
		cfg:    cfg,
		cancel: cancel,
	}

	service.Ping = run.ping
	err = service.Run()
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	return &Service{
		Container: service.Container,
		HTTPPort:  run.http,
		SSHPort:   run.ssh,
	}, nil
}
