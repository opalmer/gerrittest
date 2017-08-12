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
	"path/filepath"
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

// GenerateSSHKeys will generate and return an SSH key pair.
func GenerateSSHKeys() (ssh.PublicKey, *rsa.PrivateKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	public, err := ssh.NewPublicKey(&private.PublicKey)
	return public, private, err
}

// ReadSSHKeys will read the provided private key and return the public and
// private portions.
func ReadSSHKeys(path string) (ssh.PublicKey, *rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	private, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, nil, err
	}

	public, err := ssh.NewPublicKey(&private.PublicKey)
	return public, private, nil
}

// WritePrivateKey will take a private key and write out the public
// and private portions to disk.
func WritePrivateKey(key *rsa.PrivateKey, private string) error {
	if err := os.MkdirAll(filepath.Dir(private), 0700); err != nil {
		return err
	}

	// Write private key to disk
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	privateFile, err := os.OpenFile(private, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	return pem.Encode(privateFile, privateKeyPEM)
}
