package gerrittest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
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
		"svc":  "gerrittest",
		"cmp":  "SSHPort",
		"path": config.PrivateKeyPath,
	})

	_, private, err := ReadSSHKeys(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	sshClient, err := ssh.Dial(
		"tcp", fmt.Sprintf("%s:%d", port.Address, port.Public),
		&ssh.ClientConfig{
			User:            config.Username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(private)},
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
	return private, nil
}

// ReadSSHKeys will read the provided private key and return the public and
// private portions.
func ReadSSHKeys(path string) (ssh.PublicKey, ssh.Signer, error) {
	logger := log.WithFields(log.Fields{
		"cmp":   "ssh",
		"phase": "read-ssh-key",
		"path":  path,
	})

	logger.WithField("action", "read-file").Debug()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	logger.WithField("action", "parse-key").Debug()
	signer, err := ssh.ParsePrivateKey(data)
	return signer.PublicKey(), signer, err
}

// WriteRSAKey will take a private key and write out the public
// and private portions to disk.
// nolint: interfacer
func WriteRSAKey(key *rsa.PrivateKey, file *os.File) error {
	logger := log.WithFields(log.Fields{
		"cmp":   "ssh",
		"phase": "write-ssh-key",
	})

	logger.WithField("action", "validate").Debug()
	if err := key.Validate(); err != nil {
		return err
	}

	logger.WithField("action", "encode").Debug()
	defer file.Close() // nolint: errcheck
	return pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// SSHKey contains information about an ssh key.
type SSHKey struct {
	Public    ssh.PublicKey `json:"-"`
	Private   ssh.Signer    `json:"-"`
	Path      string        `json:"path"`
	Generated bool          `json:"generated"`
	Default   bool          `json:"default"`
}

// Remove will remove the key material from disk but only if the key was
// generated.
func (s *SSHKey) Remove() error {
	if s.Generated {
		return os.Remove(s.Path)
	}
	return nil
}

// LoadSSHKey loads an SSH key from disk and returns a *SSHKey struct.
func LoadSSHKey(path string) (*SSHKey, error) {
	public, signer, err := ReadSSHKeys(path)
	if err != nil {
		return nil, err
	}
	return &SSHKey{
		Public:    public,
		Private:   signer,
		Path:      path,
		Generated: false,
		Default:   true,
	}, nil
}

// NewSSHKey will generate and return an *SSHKey.
// TODO add tests
func NewSSHKey() (*SSHKey, error) {
	key, err := GenerateRSAKey()
	if err != nil {
		return nil, err
	}
	file, err := ioutil.TempFile(
		"", fmt.Sprintf("%s-id_rsa-", ProjectName))
	if err != nil {
		return nil, err
	}
	defer file.Name() // nolint: errcheck

	buffer := &bytes.Buffer{}
	if err := pem.Encode(buffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, buffer); err != nil {
		return nil, err
	}

	return &SSHKey{
		Private:   signer,
		Public:    signer.PublicKey(),
		Path:      file.Name(),
		Generated: true,
		Default:   true,
	}, nil
}
