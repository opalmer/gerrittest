package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
