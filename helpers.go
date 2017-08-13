package gerrittest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

var (
	// ErrExpectedCreated may be returned when we're attemping to create
	// a user and fail in the process of doing so.
	ErrExpectedCreated = errors.New("Expected 201 Created")
)

// Helpers provides a small set of utility functions
// for interacting with Gerrit inside the docker
// container. This has few external dependencies so it's
// a decent way of interacting with Gerrit when testing.
type Helpers struct {
	HTTP   *dockertest.Port
	SSH    *dockertest.Port
	log    *log.Entry
	client *http.Client
}

// GetURL returns the full url by combining port information with a tail.
func (h *Helpers) GetURL(tail string) string {
	return fmt.Sprintf("http://%s:%d%s", h.HTTP.Address, h.HTTP.Public, tail)
}

// CreateSSHKeyPair generates a public and private SSH key pair then returns
// their paths. Original code from https://stackoverflow.com/a/34347463.
func (h *Helpers) CreateSSHKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}

	privateKeyFile, err := ioutil.TempFile("", "SSHPort-rsa-")
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
func (h *Helpers) CreateAdmin() (string, string, error) {
	url := h.GetURL("/login/#/?account_id=1000000")
	logger := h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "create-admin",
	})
	logger.Debug()
	_, err := h.client.Get(url)
	if err != nil {
		logger.WithError(err).Error()
		return "", "", err
	}

	return "admin", "secret", h.CheckHTTPLogin("admin", "secret")
}

// AddPublicKeyFromPath adds the public key from the provided path to the provided user.
func (h *Helpers) AddPublicKeyFromPath(user string, password string, publicKeyPath string) error {
	file, err := os.Open(publicKeyPath)
	if err != nil {
		return err
	}
	url := h.GetURL("/a/accounts/self/sshkeys")
	h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "add-public-key",
		"user":  user,
		"path":  publicKeyPath,
	}).Debug()
	request, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)
	response, err := h.client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusCreated {
		h.log.WithError(ErrExpectedCreated).WithField("response", response).Error()
		return ErrExpectedCreated
	}
	return nil
}

// CheckHTTPLogin attempts to login as the requested user.
func (h *Helpers) CheckHTTPLogin(user string, password string) error {
	url := h.GetURL("/a/accounts/self")
	h.log.WithFields(log.Fields{
		"url":   url,
		"phase": "check-HTTP-login",
		"user":  user,
	}).Debug()
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)
	_, err = h.client.Do(request)
	return err
}

// GetSSHClient returns an SSHPort client for connecting to Gerrit docker instances.
func (h *Helpers) GetSSHClient(user *User) (*SSHClient, error) {
	return NewSSHClient(user.Login, user.PrivateKey, h.SSH)
}

// NewHelpers returns a *Helpers struct
func NewHelpers(httpPort *dockertest.Port, sshPort *dockertest.Port) *Helpers {
	jar, _ := cookiejar.New(nil)
	return &Helpers{
		HTTP: httpPort, SSH: sshPort,
		client: &http.Client{
			Jar: jar,
		},
		log: log.WithFields(log.Fields{
			"svc": "gerrittest",
			"cmp": "helpers",
		}),
	}
}
