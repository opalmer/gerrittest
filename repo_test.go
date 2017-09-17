package gerrittest

import (
	"os"
	"path/filepath"
	"strings"

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

func (s *RepoTest) TestDestroy(c *C) {
	repo := s.newRepository(c)
	c.Assert(repo.Destroy(), IsNil)
	_, err := os.Stat(repo.Root)
	c.Assert(os.IsNotExist(err), Equals, true)
}
