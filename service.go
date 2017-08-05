package gerrittest

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-gerrit"
	"github.com/opalmer/dockertest"
)

const (
	// ExportedHTTPPort is the port exported by the docker container
	// where the http service is running.
	ExportedHTTPPort = 8080

	// ExportedSSHPort is the port exported by the docker container
	// where the ssh service is running.
	ExportedSSHPort = 29418
)

// Service used to store and information about the running service.
type Service struct {
	cfg    *Config
	svc    *dockertest.Service
	log    *log.Entry
	URL    string
	Client *gerrit.Client
}

func (s *Service) ping(input *dockertest.PingInput) error {
	logger := s.log.WithField("phase", "ping")
	portHTTP, err := s.svc.Container.Port(ExportedHTTPPort)
	if err != nil {
		return err
	}

	// Wait for HTTP to be listening before moving forward.
	s.URL = fmt.Sprintf("http://%s:%d", portHTTP.Address, portHTTP.Public)
	entry := logger.WithField("action", "wait-http")
	start := time.Now()
	for {
		entry.Debug()
		if _, err := http.Get(s.URL); err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	entry.WithField("duration", time.Since(start)).Info()

	portSSH, err := s.svc.Container.Port(ExportedSSHPort)
	if err != nil {
		return err
	}
	helpers := NewHelpers(portHTTP, portSSH)

	var adminUser string
	var adminPassword string
	if s.cfg.CreateAdmin {
		username, password, pubKey, privKey, err := helpers.CreateAdmin()
		adminUser = username
		adminPassword = password
		_ = pubKey
		_ = privKey

		if err != nil {
			return err
		}
	}

	_ = adminUser
	_ = adminPassword
	fmt.Println(helpers.CreateSSHKeyPair())

	return nil
}

// Run wraps the Run function provided by dockertest and hooks up
// our own ping function.
func (s *Service) Run() error {
	s.log.WithField("phase", "run").Debug()
	s.svc.Ping = s.ping
	if err := s.svc.Run(); err != nil {
		return err
	}
	return nil
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
