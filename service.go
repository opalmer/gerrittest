package gerrittest

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
)

const (
	// ExportedHTTPPort is the port exported by the docker container
	// where the HTTPPort service is running.
	ExportedHTTPPort = 8080

	// ExportedSSHPort is the port exported by the docker container
	// where the SSHPort service is running.
	ExportedSSHPort = 29418
)

// User represents a single user for connecting to Gerrit.
type User struct {
	Login      string
	Password   string
	PrivateKey string
}

// Service used to store and information about the running service.
type Service struct {
	cfg     *Config
	svc     *dockertest.Service
	log     *log.Entry
	URL     string
	Admin   *User
	Helpers *Helpers
}

// Close will terminate the running Gerrit container. This should be
// called when you're done testing.
func (s *Service) Close() error {
	return s.svc.Terminate()
}

func (s *Service) ping(input *dockertest.PingInput) error {
	logger := s.log.WithField("phase", "ping")
	portHTTP, err := s.svc.Container.Port(ExportedHTTPPort)
	if err != nil {
		return err
	}

	// Wait for HTTP to be listening before moving forward.
	s.URL = fmt.Sprintf("HTTPPort://%s:%d", portHTTP.Address, portHTTP.Public)
	entry := logger.WithField("action", "wait-for-HTTPPort")
	start := time.Now()
	entry.Debug()
	for {

		if _, err := http.Get(s.URL); err != nil {
			time.Sleep(time.Microsecond * 100)
			continue
		}
		break
	}
	entry.WithField("duration", time.Since(start)).Debug()

	portSSH, err := s.svc.Container.Port(ExportedSSHPort)
	if err != nil {
		return err
	}
	s.Helpers = NewHelpers(portHTTP, portSSH)

	var adminUser string
	var adminPassword string
	if s.cfg.CreateAdmin {
		username, password, pubKey, privKey, err := s.Helpers.CreateAdmin()
		adminUser = username
		adminPassword = password
		if err != nil {
			return err
		}
		if err := s.Helpers.AddPublicKey(adminUser, adminPassword, pubKey); err != nil {
			return err
		}
		s.Admin = &User{
			Login:      adminUser,
			Password:   adminPassword,
			PrivateKey: privKey,
		}
		client, err := s.Helpers.GetSSHClient(s.Admin)
		if err != nil {
			return err
		}
		if err := client.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Run wraps the Run function provided by dockertest and hooks up
// our own ping function.
func (s *Service) Run() (*User, *Helpers, error) {
	s.log.WithField("phase", "run").Debug()
	s.svc.Ping = s.ping
	if err := s.svc.Run(); err != nil {
		defer s.svc.Terminate()
		return nil, nil, err
	}
	return s.Admin, s.Helpers, nil
}

// NewService constructs and returns a dockertest.Service struct.
func NewService(client *dockertest.DockerClient, cfg *Config) *Service {
	input := dockertest.NewClientInput(cfg.Image)
	input.Ports.Add(&dockertest.Port{
		Private:  ExportedHTTPPort,
		Public:   cfg.PortHTTP,
		Protocol: dockertest.ProtocolTCP,
	})
	input.Ports.Add(&dockertest.Port{
		Private:  ExportedSSHPort,
		Public:   cfg.PortHTTP,
		Protocol: dockertest.ProtocolTCP,
	})
	svc := client.Service(input)
	svc.Name = "gerrittest"
	return &Service{
		svc: svc, cfg: cfg,
		log: log.WithFields(log.Fields{
			"svc": "gerrittest",
			"cmp": "service",
		}),
	}
}
