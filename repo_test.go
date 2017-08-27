package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

func (s *RepoTest) TestRepository_Init_ExistingPath(c *C) {
	r := s.newRepo(c)
	c.Assert(r.Remove(), IsNil)
}

func (s *RepoTest) TestRepository_CreateRemoteFromSpec(c *C) {
	r := s.newRepo(c)
	_ = r
	// TODO - Add test for CreateRemoteFromSpec
}
