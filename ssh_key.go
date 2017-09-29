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
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

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
