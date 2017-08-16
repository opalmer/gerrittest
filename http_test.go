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
	"testing"

	"github.com/andygrunwald/go-gerrit"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

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
	io.Copy(h.requestBody, request.Body)
	io.Copy(response, h.response.Body)
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
		Client: &http.Client{Jar: NewCookieJar()},
		Prefix: fmt.Sprintf("http://%s", server.Listener.Addr()),
		User:   "admin",
	}
	return client, handler, server
}

func TestGetResponseBody(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString(")]}'\nfoobar"))
	response := &http.Response{Body: body}
	data, err := GetResponseBody(response)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string([]byte("foobar")) {
		t.Fatal()
	}
}

func TestHTTPClient_URL(t *testing.T) {
	client := HTTPClient{Prefix: "http://localhost"}
	if client.URL("/foo") != "http://localhost/foo" {
		t.Fatal()
	}
}

func TestHTTPClient_NewRequest(t *testing.T) {
	client, _, server := newClient(nil)
	server.Close()
	request, err := client.NewRequest(
		http.MethodGet, "/a/accounts/foo", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodGet {
		t.Fatal()
	}
	var body bytes.Buffer
	io.Copy(&body, request.Body)
	if body.String() != "foo" {
		t.Fatal()
	}
	if request.Header.Get("Content-Type") != "application/json" {
		t.Fatal()
	}
}

func TestHTTPClient_Do_BadCode(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	client, _, server := newClient(expected)
	defer server.Close()
	request, err := client.NewRequest(
		http.MethodPost, "/a/foo", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Do(request, http.StatusOK); err == nil {
		t.Fatal()
	}
}

func TestHTTPClient_Do(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	expected.Body.Write([]byte("hello"))
	client, _, server := newClient(expected)
	defer server.Close()
	request, err := client.NewRequest(
		http.MethodPost, "/a/foo", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	_, body, err := client.Do(request, http.StatusCreated)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "hello" {
		t.Fatal()
	}
}

func TestHTTPClient_Login(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	client, handler, server := newClient(expected)
	defer server.Close()
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}
	request := handler.Request()
	if request.URL.Path != "/login/" {
		t.Fatal()
	}
}

func TestHTTPClient_GetAccount(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	body, err := json.Marshal(&gerrit.AccountInfo{
		Name: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}
	expected.Body.Write(body)
	client, handler, server := newClient(expected)
	defer server.Close()
	info, err := client.GetAccount()
	if err != nil {
		t.Fatal(err)
	}
	request := handler.Request()
	if request.URL.Path != "/a/accounts/self" {
		t.Fatal()
	}
	if info.Name != "foobar" {
		t.Fatal()
	}
}

func TestHTTPClient_GeneratePassword(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	body, err := json.Marshal(&gerrit.AccountInfo{
		Name: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}
	expected.Body.Write(body)
	client, handler, server := newClient(expected)
	defer server.Close()
	if _, err := client.GeneratePassword(); err != nil {
		t.Fatal(err)
	}

	request := handler.Request()
	if request.URL.Path != "/a/accounts/self/password.http" {
		t.Fatal()
	}
	if request.Method != http.MethodPut {
		t.Fatal()
	}
}

func TestHTTPClient_SetPassword(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusOK
	client, handler, server := newClient(expected)
	defer server.Close()
	if err := client.SetPassword("foobar"); err != nil {
		t.Fatal(err)
	}

	request := handler.Request()
	if request.URL.Path != "/a/accounts/self/password.http" {
		t.Fatal()
	}
	if request.Method != http.MethodPut {
		t.Fatal()
	}

	body := handler.RequestBody()
	if body != `{"http_password":"foobar"}` {
		t.Fatal(body)
	}
}

func TestHTTPClient_InsertPublicKey(t *testing.T) {
	expected := httptest.NewRecorder()
	expected.Code = http.StatusCreated
	client, handler, server := newClient(expected)
	defer server.Close()

	private, err := GenerateRSAKey()
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	public := signer.PublicKey()
	if err := client.InsertPublicKey(public); err != nil {
		t.Fatal(err)
	}

	request := handler.Request()
	if request.URL.Path != "/a/accounts/self/sshkeys" {
		t.Fatal()
	}
	if request.Method != http.MethodPost {
		t.Fatal()
	}
	if request.Header.Get("Content-Type") != "plain/text" {
		t.Fatal()
	}

	body := handler.RequestBody()
	if body != string(bytes.TrimSpace(ssh.MarshalAuthorizedKey(public))) {
		t.Fatal(body)
	}
}

func TestNewHTTPClient(t *testing.T) {
	service := &Service{
		HTTPPort: &dockertest.Port{
			Address: "foobar",
			Public:  8080,
		},
	}
	client := NewHTTPClient(service, "admin")
	if client.Prefix != fmt.Sprintf(
		"http://%s:%d", service.HTTPPort.Address, service.HTTPPort.Public) {
		t.Fatal(client.Prefix)
	}
}
