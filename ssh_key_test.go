package gerrittest

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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

func (s *SSHKeyTest) TestLoadSSHKey(c *C) {
	key := s.generateKey(c)
	fileA := s.writeKey(c, key)
	loaded, err := LoadSSHKey(fileA)
	c.Assert(err, IsNil)
	c.Assert(loaded.Path, Equals, fileA)
	c.Assert(loaded.Private, NotNil)
	c.Assert(loaded.Public, NotNil)
	c.Assert(loaded.Generated, Equals, false)
}

func (s *SSHKeyTest) TestNewSSHKey(c *C) {
	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	c.Assert(key.Default, Equals, true)
	key.Default = false
	c.Assert(key.Remove(), IsNil)
	c.Assert(key.Generated, Equals, true)
}

func (s *SSHKeyTest) TestSSHKeyRemove(c *C) {
	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	key.Default = false
	c.Assert(key.Remove(), IsNil)
	_, err = os.Stat(key.Path)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *SSHKeyTest) TestSSHKeyRemoveDoesNotRemove(c *C) {
	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	key.Generated = false
	c.Assert(key.Remove(), IsNil)
	_, err = os.Stat(key.Path)
	c.Assert(err, IsNil)
}

func (s *SSHKeyTest) TestSSHKeyString(c *C) {
	key, err := NewSSHKey()
	c.Assert(err, IsNil)
	key.Default = false
	c.Assert(key.Remove(), IsNil)
	c.Assert(key.String(), Equals, fmt.Sprintf(
		"SSHKey{path: %s, generated: %t, default: %t}",
		key.Path, key.Generated, key.Default))
}
