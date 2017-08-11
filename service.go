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

// Service used to store and information about the running service.
type Service struct {
	cfg     *Config
	log     *log.Entry
	URL     string
	Service *dockertest.Service
	Admin   *User
	Helpers *Helpers
}

// Close will terminate the running Gerrit container. This should be
// called when you're done testing.
func (s *Service) Close() error {
	if s.cfg.Keep {
		return nil
	}
	return s.Service.Terminate()
}

func (s *Service) ping(input *dockertest.PingInput) error {
	logger := s.log.WithField("phase", "ping")
	portHTTP, err := s.Service.Container.Port(ExportedHTTPPort)
	if err != nil {
		return err
	}

	// Wait for HTTP to be listening before moving forward.
	s.URL = fmt.Sprintf("http://%s:%d", portHTTP.Address, portHTTP.Public)
	entry := logger.WithField("action", "wait-for-http")
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

	portSSH, err := s.Service.Container.Port(ExportedSSHPort)
	if err != nil {
		if s.cfg.Keep {
			logger.WithError(err).Warn()
			err = nil
		}
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
			if s.cfg.Keep {
				logger.WithError(err).Warn()
				err = nil
			}
			return err
		}
		if err := s.Helpers.AddPublicKey(adminUser, adminPassword, pubKey); err != nil {
			if s.cfg.Keep {
				logger.WithError(err).Warn()
				err = nil
			}
			return err
		}
		s.Admin = &User{
			Login:      adminUser,
			Password:   adminPassword,
			PrivateKey: privKey,
		}
		client, err := s.Helpers.GetSSHClient(s.Admin)
		if err != nil {
			if s.cfg.Keep {
				logger.WithError(err).Warn()
				err = nil
			}
			return err
		}
		if err := client.Close(); err != nil {
			if s.cfg.Keep {
				logger.WithError(err).Warn()
				err = nil
			}
			return err
		}
	}

	return nil
}

// Run wraps the Run function provided by dockertest and hooks up
// our own ping function.
func (s *Service) Run() (*User, *Helpers, error) {
	s.log.WithField("phase", "run").Debug()
	s.Service.Ping = s.ping
	if err := s.Service.Run(); err != nil {
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
		Public:   cfg.PortSSH,
		Protocol: dockertest.ProtocolTCP,
	})
	svc := client.Service(input)
	svc.Name = "gerrittest"
	return &Service{
		Service: svc, cfg: cfg,
		log: log.WithFields(log.Fields{
			"svc": "gerrittest",
			"cmp": "service",
		}),
	}
}
