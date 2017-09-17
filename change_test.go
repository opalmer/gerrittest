package gerrittest

import (
	"testing"

	. "gopkg.in/check.v1"
)

type ChangeTest struct {
	gerrit *Gerrit
	change *Change
}

var _ = Suite(&ChangeTest{})

func (s *ChangeTest) SetUpSuite(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}

	gerrit, err := New(NewConfig())
	c.Assert(err, IsNil)
	s.gerrit = gerrit
}

func (s *ChangeTest) TearDownSuite(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}
	c.Assert(s.gerrit.Destroy(), IsNil)
}

func (s *ChangeTest) SetUpTest(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}

	change, err := s.gerrit.CreateChange("", "testing")
	c.Assert(err, IsNil)
	s.change = change
}

func (s *ChangeTest) TearDownTest(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}
	c.Assert(s.change.Destroy(), IsNil)
}

func (s *ChangeTest) TestAdd(c *C) {

}

//	c.Assert(s.change.Remove(), IsNil)
//	_, err := os.Stat(s.change.cfg)
//	c.Assert(os.IsNotExist(err), Equals, true)
//	s.change = nil
//}
//
//func (s *ChangeTest) TestAddFile(c *C) {
//	c.Assert(s.change.Add("foo.txt", 0600, []byte("hello")), IsNil)
//	c.Assert(s.change.AmendAndPush(), IsNil)
//}
//
//func (s *ChangeTest) TestRemoveFile(c *C) {
//	c.Assert(s.change.Add("foo.txt", 0600, []byte("hello")), IsNil)
//	c.Assert(s.change.AmendAndPush(), IsNil)
//	c.Assert(s.change.Remove("foo.txt"), IsNil)
//	c.Assert(s.change.AmendAndPush(), IsNil)
//}
//
//func (s *ChangeTest) TestApplyLabel(c *C) {
//	c.Assert(s.change.Add("foo.txt", 0600, []byte("hello")), IsNil)
//	c.Assert(s.change.AmendAndPush(), IsNil)
//	info, err := s.change.ApplyLabel("", CodeReviewLabel, 2)
//	c.Assert(err, IsNil)
//	c.Assert(info.Labels[CodeReviewLabel], Equals, 2)
//}
