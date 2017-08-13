package gerrittest

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"bytes"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-gerrit"
)

// HTTPClient is a simple client for talking to Gerrit within a
// container. This is not intended as a replacement for go-gerrit.
// Instead, it's intended to get validate that Gerrit is setup
// correctly and then perform the final steps to get it ready for
// testing.
type HTTPClient struct {
	Client *http.Client
	Prefix string
	log    *log.Entry
}

// URL concatenates the prefix and the given tai.
func (h *HTTPClient) URL(tail string) string {
	return h.Prefix + tail
}

// Login will attempt to hit /login/ as the given user.
func (h *HTTPClient) Login(username string) error {
	url := h.URL("/login/")
	h.log.WithFields(log.Fields{
		"user":   username,
		"method": "GET",
		"url":    url,
	})
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("X-User", username)
	if err != nil {
		return err
	}

	response, err := h.Client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"Response code %d != %d", response.StatusCode, http.StatusOK)
	}
	return nil
}

// GetAccount will return information about the
func (h *HTTPClient) GetAccount(username string) error {
	url := h.URL(fmt.Sprintf("/a/accounts/%s", username))
	h.log.WithFields(log.Fields{
		"user":   username,
		"method": "GET",
		"url":    url,
	})
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := h.Client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"Response code %d != %d", response.StatusCode, http.StatusOK)
	}
	return err
}

// GeneratePassword generates and returns the account password. Note, this
// only works for the current account (the one which set the cookie
// in GetAccount())
func (h *HTTPClient) GeneratePassword() (string, error) {
	url := h.URL("/a/accounts/self/password.http")
	h.log.WithFields(log.Fields{
		"method": "PUT",
		"url":    url,
	})
	request, err := http.NewRequest("PUT", url, nil)
	request.Header.Add("X-User", "admin")
	if err != nil {
		return "", err
	}
	response, err := h.Client.Do(request)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"Response code %d != %d", response.StatusCode, http.StatusOK)
	}
	defer response.Body.Close()

	data := &bytes.Buffer{}
	_, err = io.Copy(data, response.Body)
	if err != nil {
		return "", err
	}
	return string(gerrit.RemoveMagicPrefixLine(data.Bytes())), nil
}

// NewHTTPClient takes a *Service struct and returns an *HTTPClient. No
// validation to ensure the service is actually running is performed.
func NewHTTPClient(service *Service) (*HTTPClient, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: nil})
	if err != nil {
		return nil, err
	}
	return &HTTPClient{
		Client: &http.Client{Jar: jar},
		Prefix: fmt.Sprintf(
			"http://%s:%d", service.HTTPPort.Address, service.HTTPPort.Public),
		log: log.WithField("cmp", "http-client"),
	}, nil
}
