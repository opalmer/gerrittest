package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v4"
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

func (s *RepoTest) TestRepository_Remove_Error(c *C) {
	r := &Repository{}
	c.Assert(r.Remove(), ErrorMatches, ErrRepositoryNotInitialized.Error())
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
	defer r.Remove()
	c.Assert(r.Init(), IsNil)
	_, err := os.Stat(filepath.Join(r.Path, ".git"))
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestRepository_Init_AlreadyCalled(c *C) {
	r := &Repository{}
	defer r.Remove()
	c.Assert(r.Init(), IsNil)
	r.User = ""
	c.Assert(r.Init(), IsNil)
	c.Assert(r.User, Equals, "")
}

func (s *RepoTest) TestRepository_Init_ExistingPath(c *C) {
	r := s.newRepo(c)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_CreateRemoteFromSpec(c *C) {
	r := s.newRepo(c)
	defer r.Remove()
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
}

func (s *RepoTest) TestRepository_Add_RepositoryNotInitialized(c *C) {
	r := &Repository{}
	c.Assert(r.Add(""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_Add(c *C) {
	r := s.newRepo(c)
	defer r.Remove()
	c.Assert(ioutil.WriteFile(filepath.Join(r.Path, "test.txt"), []byte("hello"), 0600), IsNil)
	c.Assert(r.Add("test.txt"), IsNil)
}
