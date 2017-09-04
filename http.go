package gerrittest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-gerrit"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

// getResponseBody returns the body of the given response as bytes with the
// magic prefix removed.
func getResponseBody(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return gerrit.RemoveMagicPrefixLine(body), response.Body.Close()
}

// HTTPClient is a simple client for talking to Gerrit within a
// container. This is not intended as a replacement for go-gerrit.
// Instead, it's intended to get validate that Gerrit is setup
// correctly and then perform the final steps to get it ready for
// testing.
type HTTPClient struct {
	Client   *http.Client
	Prefix   string
	Username string
	Password string
}

// URL concatenates the prefix and the given tai.
func (h *HTTPClient) URL(tail string) string {
	return h.Prefix + tail
}

// NewRequest constructs a new http.Request, sets the proper headers and then
// logs the request.
func (h *HTTPClient) NewRequest(method string, tail string, body []byte) (*http.Request, error) {
	requestURL := h.URL(tail)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	request, err := http.NewRequest(method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	if h.Username != "" && h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}

	// If the url is not prefixed with /a/ then assume we're relying
	// on X-User to tell Gerrit to trust our request. In all other cases
	// the cookie Gerrit gives us back will be relies
	if !strings.HasPrefix(tail, "/a/") {
		request.Header.Add("X-User", h.Username)
	}

	for _, cookie := range h.Client.Jar.Cookies(&url.URL{Host: "localhost"}) {
		request.AddCookie(cookie)
		if cookie.Name == "XSRF_TOKEN" {
			request.Header.Set("X-Gerrit-Auth", cookie.Value)
		}
	}

	log.WithFields(log.Fields{
		"action": "request",
		"method": method,
		"url":    requestURL,
		"body":   string(body),
	}).Debug()

	return request, nil
}

// Do performs the request using the internal http client.
func (h *HTTPClient) Do(request *http.Request, expectedCode int) (*http.Response, []byte, error) {
	logger := log.WithFields(log.Fields{
		"action": "response",
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
		return response, nil, err
	}

	body, err := getResponseBody(response)
	if err != nil {
		return nil, nil, err
	}

	logger = logger.WithFields(log.Fields{
		"duration": time.Since(start),
		"status":   response.StatusCode,
	})

	if expectedCode == 0 {
		expectedCode = response.StatusCode
	}
	if response.StatusCode != expectedCode {
		logger.WithField("body", strings.TrimSpace(string(body))).Warn()
		return response, body, fmt.Errorf(
			"Response code %d != %d", response.StatusCode, expectedCode)
	}
	logger.Debug()
	return response, body, err
}

// Login will attempt to hit /login/ as the given user.
func (h *HTTPClient) Login() error {
	request, err := h.NewRequest(http.MethodGet, "/login/", nil)
	if err != nil {
		return err
	}
	_, _, err = h.Do(request, http.StatusOK)
	return err
}

// Gerrit will return a *gerrit.Gerrit client. Note, the username
// and password must already be set and basic validation to ensure
// the client is setup properly is performed.
func (h *HTTPClient) Gerrit() (*gerrit.Client, error) {
	if h.Username == "" || h.Password == "" {
		return nil, errors.New("Username and password required")
	}
	parsed, err := url.Parse(h.Prefix)
	if err != nil {
		return nil, err
	}
	client, err := gerrit.NewClient(fmt.Sprintf(
		"%s://%s:%s@%s", parsed.Scheme, h.Username, h.Password,
		parsed.Host), nil)
	if err != nil {
		return nil, err
	}

	if _, _, err := client.Accounts.GetAccount("self"); err != nil {
		return nil, err
	}
	return client, nil
}

// GeneratePassword generates and returns the account password. Note, this
// only works for the current account (the one which set the cookie
// in GetAccount())
func (h *HTTPClient) GeneratePassword() (string, error) {
	body, err := json.Marshal(&gerrit.HTTPPasswordInput{Generate: true})
	if err != nil {
		return "", err
	}

	request, err := h.NewRequest(http.MethodPut, "/a/accounts/self/password.http", body)
	if err != nil {
		return "", err
	}

	_, responseBody, err := h.Do(request, http.StatusOK)
	if err != nil {
		return "", err
	}

	// The generated password includes quotes, the below code removes
	// those quotes.
	output := strings.TrimSpace(string(responseBody))
	return output[1 : len(output)-1], nil
}

// SetPassword sets the http password to the given value.
func (h *HTTPClient) SetPassword(password string) error {
	body, err := json.Marshal(&gerrit.HTTPPasswordInput{HTTPPassword: password})
	if err != nil {
		return err
	}

	request, err := h.NewRequest(
		http.MethodPut, "/a/accounts/self/password.http", body)
	if err != nil {
		return err
	}

	_, _, err = h.Do(request, http.StatusOK)
	return err
}

// InsertPublicKey will insert the provided public key.
func (h *HTTPClient) InsertPublicKey(key ssh.PublicKey) error {
	request, err := h.NewRequest(
		http.MethodPost, "/a/accounts/self/sshkeys",
		bytes.TrimSpace(ssh.MarshalAuthorizedKey(key)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "plain/text")
	_, _, err = h.Do(request, http.StatusCreated)
	return err
}

// NewHTTPClient takes a *Service struct and returns an *HTTPClient. No
// validation to ensure the service is actually running is performed.
func NewHTTPClient(username string, password string, port *dockertest.Port) (*HTTPClient, error) {
	if username == "" {
		return nil, errors.New("Username not provided")
	}
	return &HTTPClient{
		Client:   &http.Client{Jar: NewCookieJar()},
		Prefix:   fmt.Sprintf("http://%s:%d", port.Address, port.Public),
		Username: username,
		Password: password,
	}, nil
}
