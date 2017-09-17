package gerrittest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type RepoTest struct {
	repos []*Repository
}

var _ = Suite(&RepoTest{})

func (s *RepoTest) TearDownTest(c *C) {
	for _, repo := range s.repos {
		c.Assert(repo.Destroy(), IsNil)
	}
	s.repos = []*Repository{}
}

func (s *RepoTest) newRepository(c *C) *Repository {
	cfg := NewConfig()
	repo, err := NewRepository(cfg)
	c.Assert(err, IsNil)
	s.repos = append(s.repos, repo)
	return repo
}

func (s *RepoTest) TestNewRepository_InitCalled(c *C) {
	repo := s.newRepository(c)
	stat, err := os.Stat(filepath.Join(repo.Root, ".git"))
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, true)
}

func (s *RepoTest) TestNewRepository_InstallsCommitHook(c *C) {
	repo := s.newRepository(c)
	stat, err := os.Stat(filepath.Join(repo.Root, ".git", "hooks", "commit-msg"))
	c.Assert(err, IsNil)
	c.Assert(stat.Mode(), Equals, os.FileMode(0700))
}

func (s *RepoTest) TestNewRepository_ConfiguresRepo(c *C) {
	repo := s.newRepository(c)
	cfg := NewConfig()

	for key, value := range cfg.GitConfig {
		stdout, _, err := repo.Git([]string{"config", "--local", "--get", key})
		c.Assert(err, IsNil)
		c.Assert(strings.Contains(stdout, value), Equals, true)
	}
}

func (s *RepoTest) TestAmend_NoCommits(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.Amend(), ErrorMatches, ErrNoCommits.Error())
}

func (s *RepoTest) TestCommitAndAmend(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.Commit("hello"), IsNil)
	c.Assert(repo.Amend(), IsNil)
}

func (s *RepoTest) TestGetRemote_ErrRemoteDoesNotExist(c *C) {
	repo := s.newRepository(c)
	_, err := repo.GetRemote("test")
	c.Assert(err, ErrorMatches, ErrRemoteDoesNotExist.Error())
}

func (s *RepoTest) TestAddRemote_DoesNotExist(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.AddRemote("test", "http://localhost/"), IsNil)
	remote, err := repo.GetRemote("test")
	c.Assert(err, IsNil)
	c.Assert(remote, Equals, "http://localhost/")
}

func (s *RepoTest) TestAddRemote_ErrRemoteDiffers(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.AddRemote("test", "http://localhost/"), IsNil)
	c.Assert(repo.AddRemote("test", "http://localhost2/"), ErrorMatches, ErrRemoteDiffers.Error())
}

func (s *RepoTest) TestAddOriginFromContainer(c *C) {
	repo := s.newRepository(c)
	container := &Container{
		SSH: &dockertest.Port{
			Address: "0.0.0.0",
			Public:  2222,
		},
	}
	c.Assert(repo.AddOriginFromContainer(container, "foo"), IsNil)
	remote, err := repo.GetRemote("origin")
	c.Assert(err, IsNil)
	c.Assert(remote, Equals, "ssh://admin@0.0.0.0:2222/foo")
}

func (s *RepoTest) TestChangeID_ErrNoCommits(c *C) {
	repo := s.newRepository(c)
	_, err := repo.ChangeID()
	c.Assert(err, ErrorMatches, ErrNoCommits.Error())
}

func (s *RepoTest) TestChangeID(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.Commit("foo"), IsNil)
	change, err := repo.ChangeID()
	c.Assert(err, IsNil)
	c.Assert(len(change), Not(Equals), 0)
}

func (s *RepoTest) TestDestroy(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.Destroy(), IsNil)
	_, err := os.Stat(repo.Root)
	c.Assert(os.IsNotExist(err), Equals, true)
}
