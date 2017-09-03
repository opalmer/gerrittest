package gerrittest

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "gopkg.in/check.v1"
)

type RepoTest struct {
	paths []string
}

var _ = Suite(&RepoTest{})

func (s *RepoTest) TearDownSuite(c *C) {
	for _, path := range s.paths {
		c.Assert(os.RemoveAll(path), IsNil)
	}
}

func (s *RepoTest) addCleanupPath(path string) {
	s.paths = append(s.paths, path)
}

func (s *RepoTest) newBareRepo(c *C) *Repository {
	cfg, err := newRepositoryConfig("", "!")
	c.Assert(err, IsNil)
	s.addCleanupPath(cfg.Path)
	return &Repository{
		mtx:  &sync.Mutex{},
		init: false,
		Cfg:  cfg,
	}
}

func (s *RepoTest) newRepoPostInit(c *C) *Repository {
	repo := s.newBareRepo(c)
	c.Assert(repo.Init(), IsNil)
	return repo
}

func (s *RepoTest) TestNewRepository(c *C) {
	cfg, err := newRepositoryConfig("", "!")
	c.Assert(err, IsNil)
	_, err = NewRepository(cfg)
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestNewRepositoryConfig_PathNotProvided(c *C) {
	cfg, err := newRepositoryConfig("", "!")
	s.addCleanupPath(cfg.Path)
	c.Assert(err, IsNil)
	stat, err := os.Stat(cfg.Path)
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, true)
}

func (s *RepoTest) TestNewRepositoryConfig_PathProvided(c *C) {
	path, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	s.addCleanupPath(path)
	cfg, err := newRepositoryConfig(path, "!")
	c.Assert(err, IsNil)
	stat, err := os.Stat(cfg.Path)
	c.Assert(err, IsNil)
	c.Assert(stat.IsDir(), Equals, true)
}

func (s *RepoTest) TestNewRepositoryConfig_MissingPrivateKeyPath(c *C) {
	path, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	s.addCleanupPath(path)
	_, err = newRepositoryConfig(path, "")
	c.Assert(err, ErrorMatches, "Missing private key")
}

func (s *RepoTest) TestNewRepositoryConfig(c *C) {
	cfg, err := newRepositoryConfig("", "!")
	c.Assert(err, IsNil)
	s.addCleanupPath(cfg.Path)
	c.Assert(cfg, DeepEquals, &RepositoryConfig{
		Path:           cfg.Path,
		Ctx:            context.Background(),
		Command:        "git",
		CommandTimeout: time.Minute * 10,
		PrivateKey:     "!",
		GitConfig: map[string]string{
			"user.name":       "admin",
			"user.email":      "admin@localhost",
			"core.sshCommand": "ssh -i ! -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no",
		},
	})
}

func (s *RepoTest) TestRepository_Init_NewRepository(c *C) {
	repo := s.newBareRepo(c)
	_, err := repo.Status()
	c.Assert(err, NotNil)
	c.Assert(repo.Init(), IsNil)
	_, err = repo.Status()
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestRepository_InstallCommitHook_RepoNotInit(c *C) {
	repo := s.newBareRepo(c)
	c.Assert(
		repo.InstallCommitHook(), ErrorMatches,
		ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_InstallCommitHook(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.InstallCommitHook(), IsNil)
	_, err := os.Stat(
		filepath.Join(repo.Cfg.Path, ".git", "hooks", DefaultCommitHookName))
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestRepository_Config_RepoNotInit(c *C) {
	repo := s.newBareRepo(c)
	c.Assert(repo.Config("foo", "bar"), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_Config(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.Config("foo.bar", "1"), IsNil)
	stdout, _, err := repo.Git([]string{"config", "--list", "--global"})
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(stdout, "foo.bar"), Equals, false)
	stdout, _, err = repo.Git([]string{"config", "--list", "--local"})
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(stdout, "foo.bar"), Equals, true)
}

func (s *RepoTest) TestRepository_Configure(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.SetConfiguration(), IsNil)
}

func (s *RepoTest) TestRepository_Add_RepoNotInit(c *C) {
	repo := s.newBareRepo(c)
	c.Assert(repo.Add(""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_Add(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(ioutil.WriteFile(filepath.Join(repo.Cfg.Path, "foo.txt"), []byte("foo"), 0600), IsNil)
	c.Assert(ioutil.WriteFile(filepath.Join(repo.Cfg.Path, "bar.txt"), []byte("bar"), 0600), IsNil)
	c.Assert(repo.Add("foo.txt", "bar.txt"), IsNil)
	status, err := repo.Status()
	c.Assert(err, IsNil)
	c.Assert(status, Equals, "A  bar.txt\nA  foo.txt\n")
}

func (s *RepoTest) TestRepository_Commit_RepoNotInit(c *C) {
	repo := s.newBareRepo(c)
	c.Assert(repo.Commit(""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}

func (s *RepoTest) TestRepository_CreateRemoteFromSpec_RepoNotInit(c *C) {
	repo := s.newBareRepo(c)
	c.Assert(repo.CreateRemoteFromSpec(nil, "", ""), ErrorMatches, ErrRepositoryNotInitialized.Error())
}
