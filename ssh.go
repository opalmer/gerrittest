package gerrittest

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/opalmer/dockertest"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// SSHClient implements an SSH client for talking to
// Gerrit.
type SSHClient struct {
	log    *log.Entry
	Client *ssh.Client
}

// Close will close the SSHPort client and session.
func (s *SSHClient) Close() error {
	err := s.Client.Close()
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			// If the error occurred due to Close() being
			// called then ignore it rather than return an
			// error. SSHClient is short lived so we should
			// expect the socket to be closed when the program
			// terminates.
			if operr.Op == "close" {
				err = nil
			}
		}
	}

	return err
}

// Run executes a command over ssh.
func (s *SSHClient) Run(command string) ([]byte, []byte, error) {
	logger := s.log.WithField("cmd", command)
	session, err := s.Client.NewSession()
	if err != nil {
		logger.WithError(err).Error()
		return nil, nil, err
	}
	defer session.Close() // nolint: errcheck
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)
	if err != nil {
		logger.WithError(err).Error()
	} else {
		logger.Debug()
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

// Version returns the current version of Gerrit.
func (s *SSHClient) Version() (string, error) {
	stdout, _, err := s.Run("gerrit version")
	return strings.Split(strings.TrimSpace(string(stdout)), " ")[2], err
}

// NewSSHClient produces an *SSHClient struct and attempts to connect to
// Gerrrit.
func NewSSHClient(config *Config, port *dockertest.Port) (*SSHClient, error) {
	logger := log.WithFields(log.Fields{
		"svc": "gerrittest",
		"cmp": "SSHPort",
	})
	if len(config.SSHKeys) == 0 {
		return nil, errors.New("no ssh keys present")
	}

	for _, key := range config.SSHKeys {
		sshClient, err := ssh.Dial(
			"tcp", fmt.Sprintf("%s:%d", port.Address, port.Public),
			&ssh.ClientConfig{
				User:            config.Username,
				Auth:            []ssh.AuthMethod{ssh.PublicKeys(key.Private)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		)
		if err != nil {
			logger.WithError(err).Warn()
			continue
		}

		client := &SSHClient{
			log:    logger,
			Client: sshClient,
		}
		version, err := client.Version()
		logger.WithField("version", version).Debug()
		return client, err
	}
	return nil, errors.New("failed to connect to ssh with any key")
}
