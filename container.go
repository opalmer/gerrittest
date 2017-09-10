package gerrittest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/crewjam/errset"
	"github.com/opalmer/dockertest"
	log "github.com/sirupsen/logrus"
)

var (
	// DefaultImage defines the default docker image to use in
	// NewConfig(). This may be overridden with the $GERRITTEST_DOCKER_IMAGE
	// environment variable.
	DefaultImage = "opalmer/gerrittest:2.14.3"
)

const (
	// DefaultImageEnvironmentVar defines the environment variable NewConfig()
	// and the tests should be using to locate the default image override.
	DefaultImageEnvironmentVar = "GERRITTEST_DOCKER_IMAGE"

	// ExportedHTTPPort is the port exported by the docker container
	// where the HTTP service is running.
	ExportedHTTPPort = 8080

	// ExportedSSHPort is the port exported by the docker container
	// where the SSHPort service is running.
	ExportedSSHPort = 29418
)

func newPort(public uint16, private uint16) (*dockertest.Port, error) {
	if private != ExportedSSHPort && private != ExportedHTTPPort {
		return nil, errors.New("Unknown private port")
	}

	// If a random port has been chosen then we need to try and determine
	// one instead of letting Docker choose. If we don't, we'll be unable
	// to properly set $GERRIT_CANONICAL_URL which Gerrit uses for redirects.
	if private == ExportedHTTPPort && public == dockertest.RandomPort {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		defer listener.Close() // nolint: errcheck
		port, err := strconv.ParseUint(
			strings.Split(listener.Addr().String(), ":")[1], 10, 16)
		if err != nil {
			return nil, err
		}
		public = uint16(port)
	}

	return &dockertest.Port{
		Private:  private,
		Public:   public,
		Protocol: dockertest.ProtocolTCP,
	}, nil
}

func waitPort(ctx context.Context, port *dockertest.Port, errs chan error) {
	addr := fmt.Sprintf("%s:%d", port.Address, port.Public)
	logger := log.WithFields(log.Fields{
		"cmp":   "container",
		"phase": "port-wait",
		"addr":  addr,
	})
	logger.WithField("task", "begin").Debug()
	started := time.Now()
	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			errs <- ctx.Err()
			return
		case <-ticker.C:
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				continue
			}
			logger.WithFields(log.Fields{
				"task":    "done",
				"elapsed": time.Since(started),
			})
			errs <- conn.Close()
			return
		}
	}
}

func waitHTTP(ctx context.Context, port *dockertest.Port, errs chan error) {
	url := fmt.Sprintf("http://%s:%d/", port.Address, port.Public)
	logger := log.WithFields(log.Fields{
		"cmp":   "container",
		"phase": "wait-port",
		"url":   url,
	})
	logger.WithField("task", "begin").Debug()
	started := time.Now()
	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			errs <- ctx.Err()
			return
		case <-ticker.C:
			response, err := http.Get(url)
			if err != nil {
				continue
			}
			if response.StatusCode != http.StatusOK {
				continue
			}
			logger.WithFields(log.Fields{
				"task":    "end",
				"elapsed": time.Since(started),
			})
			errs <- nil
			return
		}
	}
}

func getDockerClientInput(http uint16, ssh uint16, image string) (*dockertest.ClientInput, error) {
	httpPort, err := newPort(http, ExportedHTTPPort)
	if err != nil {
		return nil, err
	}
	sshPort, err := newPort(ssh, ExportedSSHPort)
	if err != nil {
		return nil, err
	}

	image = GetDockerImage(image)
	input := dockertest.NewClientInput(image)
	input.Ports.Add(httpPort)
	input.Ports.Add(sshPort)
	input.AddEnvironmentVar(
		"GERRIT_CANONICAL_URL",
		fmt.Sprintf("http://127.0.0.1:%d/", httpPort.Public))
	return input, nil
}

// GetDockerImage returns the docker image to use to run the container. If "" is
// provided as the docker then we'll check $GERRITTEST_DOCKER_IMAGE first before
// returning the value defined in DefaultImage.
func GetDockerImage(image string) string {
	if image != "" {
		return image
	}
	if value, set := os.LookupEnv(DefaultImageEnvironmentVar); set {
		return value
	}
	return DefaultImage
}

// Container stores information about a Gerrit instance running inside of
// a container.
type Container struct {
	ctx    context.Context
	Docker *dockertest.DockerClient `json:"-"`
	HTTP   *dockertest.Port         `json:"http"`
	SSH    *dockertest.Port         `json:"ssh"`
	Image  string                   `json:"image"`
	ID     string                   `json:"id"`
}

// Terminate will terminate and remove the running container.
func (c *Container) Terminate() error {
	return c.Docker.RemoveContainer(c.ctx, c.ID)
}

// NewContainer will create a new container using dockertest and return
// it. If you prefer to use an existing container use one of the LoadContainer*
// functions instead. This function will not return until the container has
// started and is listening on the requested ports.
func NewContainer(parent context.Context, http uint16, ssh uint16, image string) (*Container, error) {
	image = GetDockerImage(image)
	logger := log.WithFields(log.Fields{
		"cmp": "container",
	})
	logger.WithField("image", image).Debug()

	input, err := getDockerClientInput(http, ssh, image)
	if err != nil {
		return nil, err
	}

	client, err := dockertest.NewClient()
	if err != nil {
		return nil, err
	}

	startLog := logger.WithField("phase", "service")
	service := &dockertest.Service{
		Input:  input,
		Client: client,
		Name:   "gerrittest",
		Ping: func(input *dockertest.PingInput) error {
			pingStart := time.Now()
			entry := startLog.WithField("status", "ping")
			entry.WithField("task", "begin").Debug()

			containerSSH, err := input.Container.Port(ExportedSSHPort)
			if err != nil {
				entry.WithError(err).Error()
				return err
			}

			containerHTTP, err := input.Container.Port(ExportedHTTPPort)
			if err != nil {
				entry.WithError(err).Error()
				return err
			}

			// Wait for ports to open
			errs := make(chan error, 2)
			results := errset.ErrSet{}
			go waitPort(parent, containerSSH, errs)
			go waitHTTP(parent, containerHTTP, errs)
			for i := 0; i < 2; i++ {
				results = append(results, <-errs)
			}

			entry.WithFields(log.Fields{
				"task":    "end",
				"elapsed": time.Since(pingStart),
			}).Debug()
			return results.ReturnValue()
		},
	}

	// Call run which will start the container and wait for Ping() in
	// the service above to return.
	start := time.Now()
	startLog.WithField("task", "begin").Debug()
	if err := service.Run(); err != nil {
		errs := errset.ErrSet{}
		errs = append(errs, err)
		errs = append(errs, service.Terminate())
		return nil, errs.ReturnValue()
	}
	startLog.WithFields(log.Fields{
		"task":    "end",
		"elapsed": time.Since(start),
	}).Debug()

	portSSH, err := service.Container.Port(ExportedSSHPort)
	if err != nil {
		return nil, err
	}

	portHTTP, err := service.Container.Port(ExportedHTTPPort)
	if err != nil {
		return nil, err
	}

	return &Container{
		ctx:    parent,
		Docker: client,
		SSH:    portSSH,
		HTTP:   portHTTP,
		Image:  image,
		ID:     service.Container.ID(),
	}, nil
}
