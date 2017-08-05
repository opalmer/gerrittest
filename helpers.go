package gerrittest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
	"os"
	"errors"
)

// Helpers provides a small set of utility functions
// for interacting with Gerrit inside the docker
// container. This has few external dependencies so it's
// a decent way of interacting with Gerrit when testing.
type Helpers struct {
	http *dockertest.Port
	ssh  *dockertest.Port
	log  *log.Entry
}

// GetURL returns the full url by combining port information with a tail.
func (h *Helpers) GetURL(tail string) string {
	return fmt.Sprintf("http://%s:%d%s", h.http.Address, h.http.Public, tail)
}

// CreateSSHKeyPair generates a public and private SSH key pair then returns
// their paths. Original code from https://stackoverflow.com/a/34347463.
func (h *Helpers) CreateSSHKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}

	privateKeyFile, err := ioutil.TempFile("", "ssh-rsa-")
	defer privateKeyFile.Close()
	if err != nil {
		return "", "", err
	}
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return "", "", err
	}

	// generate and write public key
	pubKeyPath := privateKeyFile.Name() + ".pub"
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	return pubKeyPath, privateKeyFile.Name(), ioutil.WriteFile(
		pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
}

// CreateAdmin will create the default administrator account.
func (h *Helpers) CreateAdmin() (string, string, string, string, error) {
	url := h.GetURL("/login/#/?account_id=1000000")
	logger := h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "create-admin",
	})
	logger.Debug()
	response, err := http.Get(url)
	if err != nil {
		logger.WithError(err).Error()
		return "", "", "", "", err
	}
	logger.WithField("code", response.StatusCode).Info()
	pubKey, privKey, err := h.CreateSSHKeyPair()
	if err != nil {
		return "", "", "", "", err
	}

	return "admin", "secret", pubKey, privKey, h.CheckHTTPLogin("admin", "secret")
}

// AddPublicKey adds the public key from the provided path to the provided user.
func (h *Helpers) AddPublicKey(user string, password string, publicKeyPath string) error {
	file, err := os.Open(publicKeyPath)
	if err != nil {
		return err
	}
	url := h.GetURL("/a/accounts/self/sshkeys")
	logger := h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "add-public-key",
		"user":  user,
		"path": publicKeyPath,
	})
	logger.Info()
	request, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)
	_, err = http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	return nil
}

// CheckHTTPLogin attempts to login as the requested user.
func (h *Helpers) CheckHTTPLogin(user string, password string) error {
	url := h.GetURL("/a/accounts/self")
	h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "check-http-login",
		"user":  user,
	}).Info()
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)
	_, err = http.DefaultClient.Do(request)
	return err
}

// CheckSSHLogin attempts to login as the requested user.
func (h *Helpers) CheckSSHLogin(user string, privKeyPath string) error {
	h.log.WithFields(log.Fields{
		"phase": "check-ssh-login",
		"user":  user,
		"key": privKeyPath,
	}).Info()
	return errors.New("Not Implemented")
}

// NewHelpers returns a *Helpers struct
func NewHelpers(http *dockertest.Port, ssh *dockertest.Port) *Helpers {
	return &Helpers{
		http: http, ssh: ssh,
		log: log.WithFields(log.Fields{
			"svc": "gerrittest",
			"cmp": "helpers",
		}),
	}
}
