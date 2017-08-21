package gerrittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	. "gopkg.in/check.v1"
	"testing"
	"context"
)

type RepoTest struct {
	repos []*Repository
}

var _ = Suite(&RepoTest{})

func (s *RepoTest) newRepo(c *C, path string) *Repository {
	repo, err := NewRepository(path)
	c.Assert(err, IsNil)
	_, err = os.Stat(filepath.Join(repo.Root, ".git"))
	c.Assert(err, IsNil)
	s.repos = append(s.repos, repo)
	return repo
}

func (s *RepoTest) TearDownTest(c *C) {
	for _, repo := range s.repos {
		c.Assert(repo.Destroy(), IsNil)
		_, err := os.Stat(repo.Root)
		c.Assert(os.IsNotExist(err), Equals, true)
	}
}

func (s *RepoTest) TestRepository_Init_rootPathNotProvided(c *C) {
	s.newRepo(c, "")
}

func (s *RepoTest) TestRepository_Init_rootDoesNotExist(c *C) {
	path, err := ioutil.TempDir("", "gerrittest-")
	c.Assert(err, IsNil)
	c.Assert(os.RemoveAll(path), IsNil)
	s.newRepo(c, path)
}

func (s *RepoTest) TestRepository_Init_rootExists(c *C) {
	path, err := ioutil.TempDir("", "gerrittest-")
	c.Assert(err, IsNil)
	s.newRepo(c, path)
}

func (s *RepoTest) addFile(c *C, commit bool) *Repository {
	repo := s.newRepo(c, "")
	input := &FileInput{Path: "foo/bar", Content: []byte("Hello world")}
	c.Assert(repo.AddFile(input), IsNil)
	path := filepath.Join(repo.Root, "foo", "bar")
	_, err := os.Stat(path)
	c.Assert(err, IsNil)
	data, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []byte("Hello world"))
	if commit {
		c.Assert(repo.Commit("some commit"), IsNil)
		stdout, _, err := repo.Run([]string{"show"})
		c.Assert(err, IsNil)
		c.Assert(strings.Contains(stdout, "some commit"), Equals, true)
	}
	return repo
}

func (s *RepoTest) TestRepository_AddFile(c *C) {
	s.addFile(c, false)
}

func (s *RepoTest) TestRepository_Commit(c *C) {
	s.addFile(c, true)
}

func (s *RepoTest) TestRepository_Amend(c *C) {
	repo := s.addFile(c, true)
	c.Assert(repo.Amend(), IsNil)
}

func (s *RepoTest) TestRepository_Integration(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}
	svc, err := Start(context.Background(), NewConfig())
	c.Assert(err, IsNil)
	defer svc.Service.Terminate() // Terminate the container when you're done.

	repo := s.addFile(c, true)
	setup := &Setup{Service: svc}
	spec, httpClient, _, err := setup.Init()
	c.Assert(err, IsNil)
	c.Assert(httpClient.CreateProject("foobar"), IsNil)
	c.Assert(repo.Configure(spec, "foobar", "master"), IsNil)
	c.Assert(repo.Push("master"), IsNil)
}
