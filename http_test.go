package gerrittest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/andygrunwald/go-gerrit"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
	. "gopkg.in/check.v1"
)

type HTTPTest struct{}

var _ = Suite(&HTTPTest{})

type testHandler struct {
	response    *httptest.ResponseRecorder
	request     *http.Request
	requestBody *bytes.Buffer
	mtx         *sync.Mutex
}

func (h *testHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	response.WriteHeader(h.response.Code)
	if _, err := io.Copy(h.requestBody, request.Body); err != nil {
		panic(err)
	}
	if _, err := io.Copy(response, h.response.Body); err != nil {
		panic(err)
	}
	outHeaders := response.Header()
	for key := range h.response.HeaderMap {
		outHeaders.Set(key, h.response.HeaderMap.Get(key))
	}
	h.request = request
}

func (h *testHandler) Request() *http.Request {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	return h.request
}

func (h *testHandler) RequestBody() string {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	return h.requestBody.String()
}

func newClient(response *httptest.ResponseRecorder) (*HTTPClient, *testHandler, *httptest.Server) {
	handler := &testHandler{
		response:    response,
		request:     nil,
		requestBody: bytes.NewBuffer(nil),
		mtx:         &sync.Mutex{},
	}
	server := httptest.NewServer(handler)
	client := &HTTPClient{
		config: &Config{},
		client: &http.Client{Jar: NewCookieJar()},
		Prefix: fmt.Sprintf("http://%s", server.Listener.Addr()),
	}
	return client, handler, server
}

func (s *HTTPTest) TestGetResponseBody(c *C) {
	body := ioutil.NopCloser(bytes.NewBufferString(")]}'\nfoobar"))
	response := &http.Response{Body: body}
	data, err := getResponseBody(response)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte("foobar"))
}

func (s *HTTPTest) TestHTTPClient_URL(c *C) {
	client := HTTPClient{Prefix: "http://localhost"}
	c.Assert(client.url("/foo"), Equals, "http://localhost/foo")
}

func (s *HTTPTest) TestHTTPClient_NewRequest(c *C) {
	client, _, server := newClient(nil)
	server.Close()
	request, err := client.newRequest(
		http.MethodGet, "/a/accounts/foo", []byte("foo"))
	c.Assert(err, IsNil)
	c.Assert(request.Method, Equals, http.MethodGet)

	body := &bytes.Buffer{}
	_, err = io.Copy(body, request.Body)
	c.Assert(err, IsNil)
	c.Assert(body.String(), Equals, "foo")
	c.Assert(request.Header.Get("Content-Type"), Equals, "application/json")
}

func (s *HTTPTest) TestHTTPClient_Do_BadCode(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	client, _, server := newClient(expected)
	defer server.Close()
	request, err := client.newRequest(
		http.MethodPost, "/a/foo", []byte("foo"))
	c.Assert(err, IsNil)
	_, _, err = client.do(request, http.StatusOK)
	c.Assert(err, NotNil)
}

func (s *HTTPTest) TestHTTPClient_Do(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	expected.Body.Write([]byte("hello"))
	client, _, server := newClient(expected)
	defer server.Close()
	request, err := client.newRequest(
		http.MethodPost, "/a/foo", []byte("foo"))
	c.Assert(err, IsNil)
	_, body, err := client.do(request, http.StatusCreated)
	c.Assert(err, IsNil)
	c.Assert(body, DeepEquals, []byte("hello"))
}

func (s *HTTPTest) TestHTTPClient_Login(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	client, handler, server := newClient(expected)
	defer server.Close()
	c.Assert(client.login(), IsNil)
	request := handler.Request()
	c.Assert(request.URL.Path, Equals, "/login/")
}

func (s *HTTPTest) TestHTTPClient_GeneratePassword(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	body, err := json.Marshal(&gerrit.AccountInfo{
		Name: "foobar",
	})
	c.Assert(err, IsNil)

	expected.Body.Write(body)
	client, handler, server := newClient(expected)
	defer server.Close()
	_, err = client.generatePassword()
	c.Assert(err, IsNil)

	request := handler.Request()
	c.Assert(request.URL.Path, Equals, "/a/accounts/self/password.http")
	c.Assert(request.Method, Equals, http.MethodPut)
}

func (s *HTTPTest) TestHTTPClient_SetPassword(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	client, handler, server := newClient(expected)
	defer server.Close()
	c.Assert(client.setPassword("foobar"), IsNil)

	request := handler.Request()
	c.Assert(request.URL.Path, Equals, "/a/accounts/self/password.http")
	c.Assert(request.Method, Equals, http.MethodPut)
	c.Assert(handler.RequestBody(), Equals, `{"http_password":"foobar"}`)
}

func (s *HTTPTest) TestHTTPClient_InsertPublicKey(c *C) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	client, handler, server := newClient(expected)
	defer server.Close()

	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	defer key.Remove() // nolint: errcheck
	client.config.SSHKeys = append(client.config.SSHKeys, key)
	c.Assert(client.insertPublicKeys(), IsNil)
	request := handler.Request()
	c.Assert(request.URL.Path, Equals, "/a/accounts/self/sshkeys")
	c.Assert(request.Method, Equals, http.MethodPost)
	c.Assert(request.Header.Get("Content-Type"), Equals, "plain/text")
	c.Assert(handler.RequestBody(), Equals, string(bytes.TrimSpace(ssh.MarshalAuthorizedKey(key.Public))))
}

func (s *HTTPTest) TestNewHTTPClient(c *C) {
	config := &Config{
		Username: "admin",
		Password: "port",
	}
	client, err := NewHTTPClient(config, &dockertest.Port{Public: 50000, Address: "foobar"})
	c.Assert(err, IsNil)
	c.Assert(client.Prefix, Equals, "http://foobar:50000")
}

func (s *HTTPTest) TestNewHTTPClient_Error(c *C) {
	config := &Config{Username: ""}
	_, err := NewHTTPClient(config, &dockertest.Port{Public: 50000, Address: "foobar"})
	c.Assert(err, ErrorMatches, "username not provided")
}
