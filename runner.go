package gerrittest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"net"

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
	HTTP      *dockertest.Port
	SSH       *dockertest.Port
}

// runner used to run Gerrit and wait for it to come up.
type runner struct {
	ctx      context.Context
	cfg      *Config
	cancel   context.CancelFunc
	log      *log.Entry
	http     *http.Client
	portHTTP *dockertest.Port
	portSSH  *dockertest.Port
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
			response, err := s.http.Get(url + "/")
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
	s.portHTTP = portHTTP
	s.portSSH = portSSH
	return err
}

// GetService takes a given configuration and returns the resulting
// dockertest.Service struct.
func GetService(cfg *Config) (*dockertest.Service, error) {
	dockerService, err := dockertest.NewClient()
	if err != nil {
		return nil, err
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
	return svc, nil
}

// NewHTTPClient returns an instance of *http.Client that has
// been configured to support the needs of gerrittest.
func NewHTTPClient() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	return &http.Client{Jar: jar}, nil
}

// Start will start Gerrit and return a struct containing information
// on how to connect.
func Start(parent context.Context, cfg *Config) (*Service, error) {
	service, err := GetService(cfg)
	if err != nil {
		return nil, err
	}

	client, err := NewHTTPClient()
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
		http:   client,
	}

	service.Ping = run.ping
	err = service.Run()
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	return &Service{
		Container: service.Container,
		HTTP:      run.portHTTP,
		SSH:       run.portSSH,
	}, nil
}
