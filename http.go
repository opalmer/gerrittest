package gerrittest

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	log "github.com/Sirupsen/logrus"
)

// HTTPClient is a simple client for talking to Gerrit
// within a container.
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
