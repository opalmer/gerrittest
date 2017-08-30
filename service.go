package gerrittest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

const (
	// ExportedHTTPPort is the port exported by the docker container
	// where the HTTP service is running.
	ExportedHTTPPort = 8080

	// ExportedSSHPort is the port exported by the docker container
	// where the SSHPort service is running.
	ExportedSSHPort = 29418
)

var (
	// WaitDelay is used as a delay to prevent us from hammering ports while
	// waiting for Gerrit to be listening or responding.
	WaitDelay = time.Millisecond * 200

	// DefaultStartTimeout is the amount of time we'll wait for the service
	// to come up in the container.
	DefaultStartTimeout = time.Minute * 5
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
	Service   *dockertest.Service
	Container *dockertest.ContainerInfo
	HTTPPort  *dockertest.Port
	SSHPort   *dockertest.Port
}

// runner used to run Gerrit and wait for it to come up.
type runner struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *Config
	log    *log.Entry
	http   *dockertest.Port
	ssh    *dockertest.Port
}

func (s *runner) waitPortOpen(port *dockertest.Port) error {
	addr := fmt.Sprintf("%s:%d", port.Address, port.Public)
	ticker := time.NewTicker(WaitDelay)
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case <-ticker.C:
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				continue
			}
			return conn.Close()
		}
	}
}

func (s *runner) waitListenHTTP(port *dockertest.Port) error {
	url := fmt.Sprintf("http://%s:%d", port.Address, port.Public)
	ticker := time.NewTicker(WaitDelay)
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
	port, err := strconv.ParseUint(
		strings.Split(listener.Addr().String(), ":")[1], 10, 16)
	return uint16(port), listener.Close()
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
	svc.Timeout = DefaultStartTimeout

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

	if err := service.Run(); err != nil {
		return nil, err
	}

	return &Service{
		Service:   service,
		Container: service.Container,
		HTTPPort:  run.http,
		SSHPort:   run.ssh,
	}, nil
}

// Setup is a struct which may be used to initialize a service to
// prepare it for use.
type Setup struct {
	Service *Service

	// Username is the name of the administrative user. If not provided, 'admin'
	// will be used.
	Username string

	// Password is the password to assign to the administrative user. If not
	// provided then a password will be generated for you.
	Password string

	// PrivateKeyPath is the path to a private key to load and insert into
	// the running instance. If not provided then a key will be generated
	// for you.
	PrivateKeyPath string
}

func (s *Setup) err(logger *log.Entry, err error) (*ServiceSpec, *HTTPClient, *SSHClient, error) {
	logger.WithError(err).Error()
	return nil, nil, nil, err
}

func (s *Setup) setPassword(client *HTTPClient) error {
	logger := log.WithField("action", "setup")

	if s.Password == "" {
		logger = logger.WithField("status", "generate-password")
		logger.Debug()
		generated, err := client.GeneratePassword()
		if err != nil {
			return err
		}
		s.Password = generated
		return nil
	}

	logger = logger.WithField("status", "set-password")
	logger.Debug()
	return client.SetPassword(s.Password)
}

func (s *Setup) getKey() (ssh.PublicKey, ssh.Signer, error) {
	logger := log.WithField("action", "setup")

	// Generate public/private key.
	if s.PrivateKeyPath == "" {
		logger = logger.WithField("status", "generate-private-key")
		logger.Debug()
		private, err := GenerateRSAKey()
		if err != nil {
			return nil, nil, err
		}
		file, err := ioutil.TempFile("", "id_rsa-")
		if err != nil {
			return nil, nil, err
		}
		if err := WriteRSAKey(private, file); err != nil {
			return nil, nil, err
		}
		signer, err := ssh.NewSignerFromKey(private)
		if err != nil {
			return nil, nil, err
		}
		s.PrivateKeyPath = file.Name()
		defer file.Close() // nolint: errcheck
		return signer.PublicKey(), signer, nil
	}

	// Load public/private key
	logger = logger.WithField("status", "load-private-key")
	logger.Debug()
	return ReadSSHKeys(s.PrivateKeyPath)
}

// Init will initialize the administrative user, its password and insert the
// ssh key.
func (s *Setup) Init() (*ServiceSpec, *HTTPClient, *SSHClient, error) {
	if s.Username == "" {
		s.Username = "admin"
	}
	logger := log.WithField("action", "setup")
	client := NewHTTPClient(s.Service, s.Username)

	// Login will create the user.
	logger = logger.WithField("status", "login")
	logger.Debug()
	if err := client.Login(); err != nil {
		return s.err(logger, err)
	}

	// Set the password for the user.
	logger = logger.WithField("status", "set-password")
	if err := s.setPassword(client); err != nil {
		return s.err(logger, err)
	}

	logger = logger.WithField("status", "get-ssh-key")
	public, _, err := s.getKey()
	if err != nil {
		return s.err(logger, err)
	}

	logger = logger.WithField("status", "insert-ssh-key")
	logger.Debug()
	if err := client.InsertPublicKey(public); err != nil {
		return s.err(logger, err)
	}

	logger = logger.WithField("status", "new-ssh-client")
	logger.Debug()
	sshClient, err := NewSSHClient(s.Username, s.PrivateKeyPath, s.Service.SSHPort)
	if err != nil {
		return s.err(logger, err)
	}

	spec := &ServiceSpec{
		Admin: &User{
			Login:      s.Username,
			Password:   s.Password,
			PrivateKey: s.PrivateKeyPath,
		},
		Container: s.Service.Container.ID(),
		SSH:       s.Service.SSHPort,
		HTTP:      s.Service.HTTPPort,
		URL:       client.Prefix,
		SSHCommand: fmt.Sprintf(
			"ssh -p %d -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no "+
				"%s@%s", s.Service.SSHPort.Public, s.PrivateKeyPath, s.Username,
			s.Service.SSHPort.Address),
	}

	return spec, client, sshClient, nil
}
