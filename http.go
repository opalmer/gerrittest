package gerrittest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-gerrit"
)

// GetResponseBody returns the body of the given response as bytes with the
// magic prefix removed.
func GetResponseBody(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}
	return gerrit.RemoveMagicPrefixLine(body), nil
}

// HTTPClient is a simple client for talking to Gerrit within a
// container. This is not intended as a replacement for go-gerrit.
// Instead, it's intended to get validate that Gerrit is setup
// correctly and then perform the final steps to get it ready for
// testing.
type HTTPClient struct {
	Client *http.Client
	Prefix string
	User   string
	log    *log.Entry
}

// URL concatenates the prefix and the given tai.
func (h *HTTPClient) URL(tail string) string {
	return h.Prefix + tail
}

// NewRequest constructs a new http.Request, sets the proper headers and then
// logs the request.
func (h *HTTPClient) NewRequest(method string, tail string, body io.Reader) (*http.Request, error) {
	requestURL := h.URL(tail)
	request, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}

	// If the url is not prefixed with /a/ then assume we're relying
	// on X-User to tell Gerrit to trust our request. In all other cases
	// the cookie Gerrit gives us back will be relies
	if !strings.HasPrefix(tail, "/a/") {
		request.Header.Add("X-User", h.User)
	}

	h.log.WithFields(log.Fields{
		"type":    "request",
		"method":  method,
		"url":     requestURL,
		"cookies": request.Cookies(),
		"headers": request.Header,
	}).Debug()
	return request, nil
}

// Do performs the request using the internal http client.
func (h *HTTPClient) Do(request *http.Request, expectedCode int) (*http.Response, error) {
	logger := h.log.WithFields(log.Fields{
		"type":   "response",
		"method": request.Method,
		"url":    request.URL,
	})
	if expectedCode != 0 {
		logger = logger.WithField("status-expected", expectedCode)
	}

	start := time.Now()
	response, err := h.Client.Do(request)
	if err != nil {
		logger.WithError(err).Error()
		return response, err
	}
	logger.WithFields(log.Fields{
		"duration": time.Since(start),
		"status":   response.StatusCode,
		"cookies":  response.Cookies(),
		"headers":  response.Header,
	}).Debug()
	if expectedCode == 0 {
		return response, err
	}
	if response.StatusCode != expectedCode {
		return response, fmt.Errorf(
			"Response code %d != %d", response.StatusCode, expectedCode)
	}
	return response, err
}

// Login will attempt to hit /login/ as the given user.
func (h *HTTPClient) Login() error {
	request, err := h.NewRequest(http.MethodGet, "/login/", nil)
	if err != nil {
		return err
	}
	_, err = h.Do(request, http.StatusOK)
	return err
}

// GetAccount will return information about the
func (h *HTTPClient) GetAccount(username string) (*gerrit.AccountInfo, error) {
	request, err := h.NewRequest(http.MethodGet, fmt.Sprintf("/a/accounts/%s", username), nil)
	if err != nil {
		return nil, err
	}
	response, err := h.Do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	body, err := GetResponseBody(response)
	if err != nil {
		return nil, err
	}
	account := &gerrit.AccountInfo{}
	return account, json.Unmarshal(body, account)
}

// GeneratePassword generates and returns the account password. Note, this
// only works for the current account (the one which set the cookie
// in GetAccount())
func (h *HTTPClient) GeneratePassword() (string, error) {
	request, err := h.NewRequest(http.MethodPut, "/a/accounts/self/password.http", nil)
	if err != nil {
		return "", err
	}
	response, err := h.Do(request, http.StatusOK)
	defer response.Body.Close()

	data := &bytes.Buffer{}
	_, err = io.Copy(data, response.Body)
	if err != nil {
		return "", err
	}
	return string(gerrit.RemoveMagicPrefixLine(data.Bytes())), err
}

// NewHTTPClient takes a *Service struct and returns an *HTTPClient. No
// validation to ensure the service is actually running is performed.
func NewHTTPClient(service *Service, username string) (*HTTPClient, error) {
	return &HTTPClient{
		Client: &http.Client{Jar: NewCookieJar()},
		Prefix: fmt.Sprintf(
			"http://%s:%d", service.HTTPPort.Address, service.HTTPPort.Public),
		User: username,
		log:  log.WithField("cmp", "http"),
	}, nil
}
