package gerrittest

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

type ChangeTest struct {
	gerrit *Gerrit
	change *Change
}

var (
	_       = Suite(&ChangeTest{})
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func generaRandomString(length int) string {
	output := make([]rune, length)
	for i := range output {
		output[i] = letters[rand.Intn(len(letters))]
	}
	return string(output)
}

func (s *ChangeTest) SetUpSuite(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}

	gerrit, err := New(NewConfig())
	if err != nil {
		c.Fatal(err)
	}
	s.gerrit = gerrit
}

func (s *ChangeTest) TearDownSuite(c *C) {
	if s.gerrit != nil {
		c.Assert(s.gerrit.Destroy(), IsNil)
	}
}

func (s *ChangeTest) SetUpTest(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}

	change, err := s.gerrit.CreateChange(
		generaRandomString(16), generaRandomString(6))
	c.Assert(err, IsNil)
	s.change = change
}

func (s *ChangeTest) TearDownTest(c *C) {
	if s.change != nil {
		c.Assert(s.change.Destroy(), IsNil)
	}
	s.change = nil
}

func (s *ChangeTest) TestDestroy(c *C) {
	c.Assert(s.change.Destroy(), IsNil)
	_, err := os.Stat(s.change.Repo.Root)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *ChangeTest) testAdd(c *C) (string, string) {
	// WARNING: `body` must be random. Because of how the commit hook generates
	// ChangeID you need to include a random body. Otherwise you can easily
	// end up with identical change ids between test runs.
	body := generaRandomString(64)
	filename := generaRandomString(15) + ".txt"
	relative := filepath.Join("foo", filename)

	c.Assert(s.change.Add(relative, 0600, body), IsNil)
	path := filepath.Join(s.change.Repo.Root, relative)
	_, err := os.Stat(path)
	c.Assert(err, IsNil)
	content, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, body)
	stdout, _, err := s.change.Repo.Git([]string{"ls-files", "--error-unmatch", relative})
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(stdout, relative), Equals, true)
	return relative, body
}

func (s *ChangeTest) TestAdd(c *C) {
	s.testAdd(c)
}

func (s *ChangeTest) TestRemove(c *C) {
	relative, _ := s.testAdd(c)
	c.Assert(s.change.Remove(relative), IsNil)
	path := filepath.Join(s.change.Repo.Root, relative)
	_, err := os.Stat(path)
	c.Assert(os.IsNotExist(err), Equals, true)
	_, stderr, err := s.change.Repo.Git([]string{"ls-files", "--error-unmatch", relative})
	c.Assert(strings.Contains(stderr, "did not match any file(s) known to git."), Equals, true)
	c.Assert(err, NotNil)
}

func (s *ChangeTest) TestPush(c *C) {
	s.TestAdd(c)
	c.Assert(s.change.Push(), IsNil)
}

func (s *ChangeTest) testApplyLabels(c *C, labels map[string][]int) {
	s.TestPush(c)
	for key, values := range labels {
		for _, value := range values {
			info, err := s.change.ApplyLabel("1", key, value)
			c.Assert(err, IsNil)
			c.Assert(info.Labels[key], Equals, value)
		}
	}
}

func (s *ChangeTest) TestApplyLabels(c *C) {
	s.testApplyLabels(c, map[string][]int{
		CodeReviewLabel: {-2, -1, 0, 1, 2},
		VerifiedLabel:   {-1, 0, 1},
	})
}

func (s *ChangeTest) TestSubmit(c *C) {
	s.testApplyLabels(c, map[string][]int{
		CodeReviewLabel: {2},
		VerifiedLabel:   {1},
	})
	info, err := s.change.Submit()
	c.Assert(err, IsNil)

	c.Assert(info.ChangeID, Equals, s.change.ChangeID)
}
