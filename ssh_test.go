package gerrittest

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
	. "gopkg.in/check.v1"
)

type SSHTest struct{}

var _ = Suite(&SSHTest{})

func (s *SSHTest) generateKey(c *C) *rsa.PrivateKey {
	private, err := GenerateRSAKey()
	c.Assert(err, IsNil)
	return private
}

func (s *SSHTest) writeKey(c *C, key *rsa.PrivateKey) string {
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	c.Assert(pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), IsNil)
	c.Assert(file.Close(), IsNil)
	return file.Name()
}
func (s *SSHTest) TestGenerateRSAKey(c *C) {
	s.generateKey(c)
}

func (s *SSHTest) TestReadSSHKeys(c *C) {
	key := s.generateKey(c)
	path := s.writeKey(c, key)
	_, private, err := ReadSSHKeys(path)
	c.Assert(err, IsNil)
	signer, err := ssh.NewSignerFromKey(key)
	c.Assert(err, IsNil)
	c.Assert(signer.PublicKey().Marshal(), DeepEquals, private.PublicKey().Marshal())
	c.Assert(os.Remove(path), IsNil)
}

func (s *SSHTest) TestWriteRSAKey(c *C) {
	key := s.generateKey(c)
	fileA := s.writeKey(c, key)
	fileB, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)

	c.Assert(WriteRSAKey(key, fileB), IsNil)
	c.Assert(fileB.Close(), NotNil) // Shouldn't be nil because WriteRSAKey closes the handle.

	a, err := ioutil.ReadFile(fileA)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadFile(fileB.Name())
	c.Assert(err, IsNil)
	c.Assert(a, DeepEquals, b)
	c.Assert(os.Remove(fileA), IsNil)
	c.Assert(os.Remove(fileB.Name()), IsNil)
}

func (s *SSHTest) TestNewSSHClientFromService(c *C) {
	if testing.Short() {
		c.Skip("-shot set")
	}

	svc, err := Start(context.Background(), NewConfig())
	c.Assert(err, IsNil)
	setup := &Setup{Service: svc}
	_, _, _, err = setup.Init()
	c.Assert(err, IsNil)
	c.Assert(svc.Service.Terminate(), IsNil)
}
