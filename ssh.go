package gerrittest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
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
	return s.Client.Close()
}

// Run executes a command over ssh.
func (s *SSHClient) Run(command string) ([]byte, []byte, error) {
	logger := s.log.WithField("cmd", command)
	session, err := s.Client.NewSession()
	if err != nil {
		logger.WithError(err).Error()
		return nil, nil, err
	}
	defer session.Close()
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
	return stdout.Bytes(), stderr.Bytes(), err
}

// Version returns the current version of Gerrit.
func (s *SSHClient) Version() (string, error) {
	stdout, _, err := s.Run("gerrit version")
	return strings.Split(strings.TrimSpace(string(stdout)), " ")[2], err
}

// NewSSHClient produces an *SSHClient struct and attempts to connect to
// Gerrrit.
func NewSSHClient(user string, privateKeyPath string, port *dockertest.Port) (*SSHClient, error) {
	logger := log.WithFields(log.Fields{
		"svc": "gerrittest",
		"cmp": "SSHPort",
	})
	data, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}

	sshClient, err := ssh.Dial(
		"tcp", fmt.Sprintf("%s:%d", port.Address, port.Public),
		&ssh.ClientConfig{
			User:            user,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	client := &SSHClient{
		log:    logger,
		Client: sshClient,
	}
	version, err := client.Version()
	logger.WithField("version", version).Debug()
	return client, err
}

// GenerateRSAKey will generate and return an SSH key pair.
func GenerateRSAKey() (*rsa.PrivateKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	if err := private.Validate(); err != nil {
		return nil, err
	}
	return private, nil
}

// ReadSSHKeys will read the provided private key and return the public and
// private portions.
func ReadSSHKeys(path string) (ssh.PublicKey, ssh.Signer, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	signer, err := ssh.ParsePrivateKey(data)
	return signer.PublicKey(), signer, err
}

// WriteRSAKey will take a private key and write out the public
// and private portions to disk.
func WriteRSAKey(key *rsa.PrivateKey, file *os.File) error {
	if err := key.Validate(); err != nil {
		return err
	}

	return pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}
