package gerrittest

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
	. "gopkg.in/check.v1"
)

type SSHKeyTest struct{}

var _ = Suite(&SSHKeyTest{})

func (s *SSHKeyTest) generateKey(c *C) *rsa.PrivateKey {
	private, err := GenerateRSAKey()
	c.Assert(err, IsNil)
	return private
}

func (s *SSHKeyTest) writeKey(c *C, key *rsa.PrivateKey) string {
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	c.Assert(pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), IsNil)
	c.Assert(file.Close(), IsNil)
	return file.Name()
}
func (s *SSHKeyTest) TestGenerateRSAKey(c *C) {
	s.generateKey(c)
}

func (s *SSHKeyTest) TestReadSSHKeys(c *C) {
	key := s.generateKey(c)
	path := s.writeKey(c, key)
	_, private, err := ReadSSHKeys(path)
	c.Assert(err, IsNil)
	signer, err := ssh.NewSignerFromKey(key)
	c.Assert(err, IsNil)
	c.Assert(signer.PublicKey().Marshal(), DeepEquals, private.PublicKey().Marshal())
	c.Assert(os.Remove(path), IsNil)
}

func (s *SSHKeyTest) TestWriteRSAKey(c *C) {
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
