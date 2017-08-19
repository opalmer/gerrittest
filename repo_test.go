package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newRepo(t *testing.T, path string) *Repository {
	repo, err := NewRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo.Root, ".git")); err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestRepository_Init_rootPathProvided(t *testing.T) {
	repo := newRepo(t, "")
	defer repo.Destroy()
}

func TestRepository_Init_rootDoesNotExist(t *testing.T) {
	path, err := ioutil.TempDir("", "gerrittest-")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
	repo := newRepo(t, path)
	defer repo.Destroy()
}

func TestRepository_Init_rootExists(t *testing.T) {
	path, err := ioutil.TempDir("", "gerrittest-")
	if err != nil {
		t.Fatal(err)
	}
	repo := newRepo(t, path)
	defer repo.Destroy()
}

func TestRepository_Destroy(t *testing.T) {
	repo := newRepo(t, "")
	if err := repo.Destroy(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(repo.Root); err == nil {
		t.Fatal()
	}
}

func TestRepository_AddFile(t *testing.T) {
	repo := newRepo(t, "")
	defer repo.Destroy()
	if err := repo.AddFile("foo/bar", []byte("Hello world"), 0600); err != nil {
		t.Fatal()
	}
	path := filepath.Join(repo.Root, "foo", "bar")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hello world" {
		t.Fatal()
	}
}

func TestRepository_Commit(t *testing.T) {
	repo := newRepo(t, "")
	defer repo.Destroy()
	if err := repo.AddFile("foo/bar", []byte("Hello world"), 0600); err != nil {
		t.Fatal()
	}
	if err := repo.Commit("some commit"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := repo.Run([]string{"show"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "some commit") {
		t.Fatal()
	}
}

func TestRepository_Amend(t *testing.T) {
	repo := newRepo(t, "")
	defer repo.Destroy()
	if err := repo.AddFile("foo/bar", []byte("Hello world"), 0600); err != nil {
		t.Fatal()
	}
	if err := repo.Commit("some commit"); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddFile("foo/bar2", []byte("Hello world"), 0600); err != nil {
		t.Fatal()
	}
	if err := repo.Amend(); err != nil {
		t.Fatal(err)
	}

}
