package gerrittest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"golang.org/x/crypto/ssh"
)

func TestGetService(t *testing.T) {
	cfg := &Config{
		PortHTTP: dockertest.RandomPort,
	}

	svc, err := GetService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if svc.Name != "gerrittest" {
		t.Fatal()
	}
	if svc.Timeout != time.Minute*10 {
		t.Fatal()
	}
	if cfg.PortHTTP == dockertest.RandomPort {
		t.Fatal()
	}
	if len(svc.Input.Environment) != 1 {
		t.Fatal()
	}
}

func TestGetRandomPort(t *testing.T) {
	port, err := GetRandomPort()
	if err != nil {
		t.Fatal(err)
	}

	// We expect nothing to be listening on the port GetRandomPort()
	// returned. If something is listening then we didn't close the port
	// before leaving GetRandomPort().
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		t.Fatal()
	}
	if conn != nil {
		t.Fatal()
	}
}

func TestStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Service.Terminate()
}

func TestRunner_waitPortOpen_cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{
		ctx: ctx,
	}

	err := run.waitPortOpen(&dockertest.Port{Address: "127.0.0.1", Public: 0})
	if err != context.Canceled {
		t.Fatal(err)
	}
}

func TestRunner_waitPortOpen_dialError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	run := runner{
		ctx: ctx,
	}
	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()

	if err := run.waitPortOpen(&dockertest.Port{Address: "127.0.0.1", Public: 65535}); err != context.Canceled {
		t.Fatal(err)
	}
}

func TestRunner_waitListenHTTP_cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{
		ctx: ctx,
	}
	if err := run.waitListenHTTP(&dockertest.Port{Address: "127.0.0.1", Public: 65535}); err != context.Canceled {
		t.Fatal(err)
	}
}

func TestRunner_waitListenHTTP_badCode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := runner{
		ctx: ctx,
	}
	count := make(chan int, 1)
	count <- 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := <-count
		current++
		count <- current
		if current < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	if err := run.waitListenHTTP(&dockertest.Port{Address: "127.0.0.1", Public: 65535}); err != context.Canceled {
		t.Fatal(err)
	}
}

func TestSetup_err(t *testing.T) {
	setup := Setup{}
	expected := errors.New("testing")
	a, b, c, err := setup.err(log.WithField("phase", "test"), expected)
	if a != nil {
		t.Fatal()
	}
	if b != nil {
		t.Fatal()
	}
	if c != nil {
		t.Fatal()
	}
	if err != expected {
		t.Fatal(err)
	}
}

func TestSetup_getKeyPath(t *testing.T) {
	key, err := GenerateRSAKey()
	if err != nil {
		t.Fatal(err)
	}
	file, err := ioutil.TempFile("", "")
	defer os.Remove(file.Name())
	generatedSigner, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteRSAKey(key, file); err != nil {
		t.Fatal(err)
	}

	setup := &Setup{PrivateKeyPath: file.Name()}
	_, signer, err := setup.getKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(signer.PublicKey().Marshal()) != string(generatedSigner.PublicKey().Marshal()) {
		t.Fatal()
	}
}

// You can start the Gerrit service using the Start() function. This only
// starts the container and returns information about the service.
func ExampleStart() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Service.Terminate() // Terminate the container when you're done.
}

// Once you've started the service you'll want to setup Gerrit inside
// the container. Running Setup.Init will cause the administrative user to
// be created, generate an http api password and insert a public key for ssh
// access.
func ExampleSetup() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Service.Terminate() // Terminate the container when you're done.

	setup := &Setup{Service: svc}
	spec, http, ssh, err := setup.Init()
	if err != nil {
		log.Fatal(err)
	}
	_ = spec
	_ = http
	_ = ssh
}
