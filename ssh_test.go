package gerrittest

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"context"

	"golang.org/x/crypto/ssh"
)

func generateKey(t *testing.T) *rsa.PrivateKey {
	private, err := GenerateRSAKey()
	if err != nil {
		t.Fatal(err)
	}
	return private
}

func writeKey(t *testing.T, key *rsa.PrivateKey) string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return file.Name()
}

func TestGenerateRSAKey(t *testing.T) {
	generateKey(t)
}

func TestReadSSHKeys(t *testing.T) {
	key := generateKey(t)
	path := writeKey(t, key)
	defer os.Remove(path)

	_, private, err := ReadSSHKeys(path)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if string(signer.PublicKey().Marshal()) != string(private.PublicKey().Marshal()) {
		t.Fatal()
	}
}

func TestWriteRSAKey(t *testing.T) {
	key := generateKey(t)
	fileA := writeKey(t, key)
	defer os.Remove(fileA)
	fileB, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileB.Name())

	if err := WriteRSAKey(key, fileB); err != nil {
		t.Fatal(err)
	}
	fileB.Close()

	a, err := ioutil.ReadFile(fileA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadFile(fileB.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatal()
	}
}

func TestNewSSHClientFromService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Service.Terminate()

	setup := &Setup{Service: svc}
	if _, _, _, err := setup.Init(); err != nil {
		t.Fatal(err)
	}
}
