package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type RepoTest struct{}

var _ = Suite(&RepoTest{})

func (s *RepoTest) newRepo(c *C) *Repository {
	path, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	_, err = git.PlainInit(path, false)
	c.Assert(err, IsNil)
	r := &Repository{Path: path}
	c.Assert(r.Init(), IsNil)
	_, err = os.Stat(filepath.Join(r.Path, ".git"))
	c.Assert(err, IsNil)
	return r
}

func (s *RepoTest) TestRepository_Remove_NoPath(c *C) {
	r := &Repository{}
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Remove_Success(c *C) {
	path, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	r := &Repository{Path: path}
	c.Assert(r.Remove(), IsNil)
	_, err = os.Stat(path)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *RepoTest) TestRepository_Init_TempPath(c *C) {
	r := &Repository{}
	c.Assert(r.Init(), IsNil)
	_, err := os.Stat(filepath.Join(r.Path, ".git"))
	c.Assert(err, IsNil)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Init_AlreadyCalled(c *C) {
	r := &Repository{}
	c.Assert(r.Init(), IsNil)
	r.User = ""
	c.Assert(r.Init(), IsNil)
	c.Assert(r.User, Equals, "")
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Init_ExistingPath(c *C) {
	r := s.newRepo(c)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_CreateRemoteFromSpec(c *C) {
	r := s.newRepo(c)
	spec := &ServiceSpec{
		Admin: &User{
			Login: "test",
		},
		SSH: &dockertest.Port{
			Public:  55555,
			Address: "127.0.0.1",
		},
	}
	c.Assert(r.CreateRemoteFromSpec(spec, "testing", "foobar"), IsNil)
	_, err := r.Repo.Remote("testing")
	c.Assert(err, IsNil)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Add_RepositoryNotInitialized(c *C) {
	r := &Repository{}
	c.Assert(r.Add(""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) add(c *C, relative string, data []byte) *Repository {
	r := s.newRepo(c)
	c.Assert(ioutil.WriteFile(filepath.Join(r.Path, relative), data, 0600), IsNil)
	c.Assert(r.Add(relative), IsNil)
	return r
}

func (s *RepoTest) TestRepository_Add(c *C) {
	r := s.add(c, "test.txt", []byte("hello"))
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Commit_RepositoryNotInitialized(c *C) {
	r := &Repository{}
	c.Assert(r.Commit(""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_Commit(c *C) {
	r := s.add(c, "test.txt", []byte("hello"))
	c.Assert(r.Commit("hello"), IsNil)
	commits, err := r.Repo.CommitObjects()
	c.Assert(err, IsNil)
	defer commits.Close()
	found := false
	c.Assert(commits.ForEach(func(commit *object.Commit) error {
		if strings.Contains(commit.Message, "hello") {
			found = true
		}
		return nil
	}), IsNil)
	c.Assert(found, Equals, true)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_Push_RepositoryNotInitialized(c *C) {
	r := &Repository{}
	c.Assert(r.Push("", ""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_Push(c *C) {
	r := s.add(c, "test.txt", []byte("hello"))
	defer r.Remove() // nolint: errcheck
	remoteRepoPath, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	defer os.RemoveAll(remoteRepoPath) // nolint: errcheck
	_, err = git.PlainInit(remoteRepoPath, false)
	c.Assert(err, IsNil)
	_, err = r.Repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{remoteRepoPath},
		Fetch: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
	})
	c.Assert(r.Commit("testing"), IsNil)
	c.Assert(r.Push("", ""), ErrorMatches, "already up-to-date")
}
