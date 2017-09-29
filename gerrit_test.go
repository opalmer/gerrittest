package gerrittest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/opalmer/dockertest"
	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

type GerritTest struct{}

var _ = Suite(&GerritTest{})

func (s *GerritTest) gerrit(c *C) *Gerrit {
	g := &Gerrit{
		Config: NewConfig(),
		log:    log.WithField("cmp", "core"),
	}
	g.Config.Username = "admin"
	return g
}

func (s *GerritTest) serverToPort(c *C, server *httptest.Server) *dockertest.Port {
	split := strings.Split(server.Listener.Addr().String(), ":")
	port, err := strconv.ParseUint(split[1], 10, 16)
	c.Assert(err, IsNil)
	return &dockertest.Port{
		Address: split[0],
		Public:  uint16(port),
	}
}

func (s *GerritTest) addSSHKey(c *C, g *Gerrit) string {
	file, err := ioutil.TempFile("", fmt.Sprintf("%s-", ProjectName))
	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	g.Config.SSHKeys = append(g.Config.SSHKeys, key)
	return file.Name()
}

func (s *GerritTest) TestNew(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}

	cfg := NewConfig()
	gerrit, err := New(cfg)
	c.Assert(err, IsNil)
	defer gerrit.Destroy() // nolint: errcheck

	file, err := ioutil.TempFile("", fmt.Sprintf("%s-", ProjectName))
	c.Assert(err, IsNil)
	path := file.Name()
	c.Assert(file.Close(), IsNil)
	defer os.Remove(path) // nolint: errcheck

	c.Assert(gerrit.WriteJSONFile(path), IsNil)
	_, err = NewFromJSON(path)
	c.Assert(err, IsNil)
	c.Assert(gerrit.Destroy(), IsNil)
}

func (s *GerritTest) TestGerrit_setupSSHKey_noPrivateKey(c *C) {
	g := s.gerrit(c)
	defer os.Remove(s.addSSHKey(c, g)) // nolint: errcheck
	c.Assert(g.setupSSHKey(), IsNil)
}

func (s *GerritTest) TestGerrit_setupHTTPClient_passwordSet(c *C) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests {
		case 0:
			w.WriteHeader(http.StatusOK)
		case 1:
			w.WriteHeader(http.StatusCreated)
		default:
			fmt.Fprint(w, "{}")
		}
		requests++
	}))

	defer ts.Close()
	g := s.gerrit(c)
	defer os.Remove(s.addSSHKey(c, g)) // nolint: errcheck
	g.HTTPPort = s.serverToPort(c, ts)
	g.Config.Password = "foo"
	c.Assert(g.setupHTTPClient(), IsNil)
}

func (s *GerritTest) TestGerrit_setupHTTPClient_errLogin(c *C) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests {
		case 0:
			w.WriteHeader(http.StatusOK)
		case 1:
			w.WriteHeader(http.StatusBadRequest)
		}
		requests++
	}))

	defer ts.Close()
	g := s.gerrit(c)
	defer os.Remove(s.addSSHKey(c, g)) // nolint: errcheck
	g.HTTPPort = s.serverToPort(c, ts)
	c.Assert(g.setupHTTPClient(), ErrorMatches, "response code 400 != 201")
}

func (s *GerritTest) TestGerrit_setupHTTPClient_generatePassword(c *C) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests {
		case 0:
			w.WriteHeader(http.StatusOK)
		case 1:
			w.WriteHeader(http.StatusCreated)
		default:
			fmt.Fprint(w, `{"password": "hello"}`)
		}
		requests++
	}))

	defer ts.Close()
	g := s.gerrit(c)
	defer os.Remove(s.addSSHKey(c, g)) // nolint: errcheck
	g.HTTPPort = s.serverToPort(c, ts)
	c.Assert(g.setupHTTPClient(), IsNil)
}

func (s *GerritTest) TestGerrit_setupHTTPClient_errGeneratePassword(c *C) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests {
		case 0:
			w.WriteHeader(http.StatusOK)
		case 1:
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		requests++
	}))

	defer ts.Close()
	g := s.gerrit(c)
	defer os.Remove(s.addSSHKey(c, g)) // nolint: errcheck
	g.HTTPPort = s.serverToPort(c, ts)
	c.Assert(g.setupHTTPClient(), ErrorMatches, "response code 400 != 200")
}

func (s *GerritTest) TestGerrit_setupHTTPClient_errUsernameNotProvided(c *C) {
	g := s.gerrit(c)
	g.Config.Username = ""
	c.Assert(g.setupHTTPClient(), ErrorMatches, "username not provided")
}
