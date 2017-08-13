package gerrittest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type HTTPTest struct{}

var _ = Suite(&HTTPTest{})

func newService(c *C) *Service {
	httpPort, err := GetRandomPort()
	c.Assert(err, IsNil)
	sshPort, err := GetRandomPort()
	c.Assert(err, IsNil)

	return &Service{
		HTTPPort: &dockertest.Port{
			Address:  "127.0.0.1",
			Public:   httpPort,
			Private:  ExportedHTTPPort,
			Protocol: dockertest.ProtocolTCP,
		},
		SSHPort: &dockertest.Port{
			Address:  "127.0.0.1",
			Public:   sshPort,
			Private:  ExportedSSHPort,
			Protocol: dockertest.ProtocolTCP,
		},
	}
}

func newClient(c *C) *HTTPClient {
	svc := newService(c)
	client, err := NewHTTPClient(svc, "foobar")
	c.Assert(err, IsNil)
	return client
}

func (s *HTTPTest) TestGetResponseBody(c *C) {
	body := ioutil.NopCloser(bytes.NewBufferString(")]}'\nfoobar"))
	response := &http.Response{Body: body}
	data, err := GetResponseBody(response)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte("foobar"))
}

func (s *HTTPTest) TestNewHTTPClient(c *C) {
	svc := newService(c)
	client, err := NewHTTPClient(svc, "foobar")
	c.Assert(err, IsNil)
	expected := &HTTPClient{
		Client: &http.Client{Jar: NewCookieJar()},
		Prefix: fmt.Sprintf(
			"http://%s:%d", svc.HTTPPort.Address, svc.HTTPPort.Public),
		User: "foobar",
		log:  log.WithField("cmp", "http"),
	}
	c.Assert(client, DeepEquals, expected)
}

func (s *HTTPTest) TestURL(c *C) {
	client := newClient(c)
	client.Prefix = "https://localhost"
	c.Assert(client.URL("/foobar/"), Equals, "https://localhost/foobar/")
}

func (s *HTTPTest) TestNewRequest(c *C) {
	client := newClient(c)
	request, err := client.NewRequest(http.MethodDelete, "/foo", nil)
	c.Assert(err, IsNil)
	expected, err := http.NewRequest(http.MethodDelete, client.URL("/foo"), nil)
	expected.Header.Add("X-User", client.User)
	c.Assert(err, IsNil)
	c.Assert(request, DeepEquals, expected)
}

func (s *HTTPTest) TestDo(c *C) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "foo")
	}))
	defer ts.Close()
	client := newClient(c)
	client.Prefix = ts.URL
	request, err := client.NewRequest(http.MethodGet, "/", nil)
	c.Assert(err, IsNil)
	response, err := client.Do(request, nil, http.StatusOK)
	c.Assert(err, IsNil)
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	c.Assert(string(body), Equals, "foo")
}
