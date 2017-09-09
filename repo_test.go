package gerrittest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opalmer/dockertest"
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

func (s *RepoTest) newConfig(c *C) *Config {
	path, err := ioutil.TempDir("", fmt.Sprintf("%s-", ProjectName))
	c.Assert(err, IsNil)
	s.addCleanupPath(path)
	file, err := ioutil.TempFile("", fmt.Sprintf("%s-", ProjectName))
	c.Assert(err, IsNil)
	private, err := GenerateRSAKey()
	c.Assert(err, IsNil)
	c.Assert(WriteRSAKey(private, file), IsNil)
	cfg := NewConfig()
	cfg.RepoRoot = path
	cfg.PrivateKeyPath = file.Name()
	return cfg
}

func (s *RepoTest) newBareRepo(c *C) *Repository {
	return &Repository{config: s.newConfig(c)}
}

func (s *RepoTest) newRepoPostInit(c *C) *Repository {
	repo := s.newBareRepo(c)
	_, _, err := repo.Git(DefaultGitCommands["init"])
	c.Assert(err, IsNil)
	return repo
}

func (s *RepoTest) TestNewRepository(c *C) {
	_, err := NewRepository(s.newConfig(c))
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestNewRepository_MissingPrivateKey(c *C) {
	cfg := s.newConfig(c)
	cfg.PrivateKeyPath = ""
	_, err := NewRepository(cfg)
	c.Assert(err, ErrorMatches, "missing private key")
}

func (s *RepoTest) TestRepository_Init_NewRepository(c *C) {
	repo := s.newBareRepo(c)
	_, err := repo.Status()
	c.Assert(err, NotNil)
	c.Assert(repo.Init(), IsNil)
	_, err = repo.Status()
	c.Assert(err, IsNil)
}

func (s *RepoTest) TestRepository_Config(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.ConfigLocal("foo.bar", "1"), IsNil)
	stdout, _, err := repo.Git([]string{"config", "--list", "--global"})
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(stdout, "foo.bar"), Equals, false)
	stdout, _, err = repo.Git([]string{"config", "--list", "--local"})
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(stdout, "foo.bar"), Equals, true)
}

func (s *RepoTest) TestRepository_Add(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(ioutil.WriteFile(filepath.Join(repo.config.RepoRoot, "foo.txt"), []byte("foo"), 0600), IsNil)
	c.Assert(ioutil.WriteFile(filepath.Join(repo.config.RepoRoot, "bar.txt"), []byte("bar"), 0600), IsNil)
	c.Assert(repo.Add("foo.txt", "bar.txt"), IsNil)
	status, err := repo.Status()
	c.Assert(err, IsNil)
	c.Assert(status, Equals, "A  bar.txt\nA  foo.txt\n")
}

func (s *RepoTest) TestRepository_AddRemote_ErrRemoteExists(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.AddRemote("foo", "bar"), IsNil)
	c.Assert(repo.AddRemote("foo", "bar"), ErrorMatches, ErrRemoteExists.Error())
}

func (s *RepoTest) TestRepository_AddRemoteFromContainer(c *C) {
	repo := s.newRepoPostInit(c)
	container := &Container{
		SSH: &dockertest.Port{
			Address: "localhost",
			Public:  5000,
		},
	}
	c.Assert(repo.AddRemoteFromContainer(
		container, "gerrittest", "foo/bar"), IsNil)
	url, err := repo.GetRemoteURL("gerrittest")
	c.Assert(err, IsNil)
	c.Assert(url, Equals, "ssh://admin@localhost:5000/foo/bar")
}

func (s *RepoTest) TestRepository_WriteFile(c *C) {
	repo := s.newRepoPostInit(c)
	c.Assert(repo.AddContent("foo/bar.txt", 0600, []byte("Hello")), IsNil)
	content, err := ioutil.ReadFile(filepath.Join(repo.config.RepoRoot, "foo", "bar.txt"))
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello")
}
