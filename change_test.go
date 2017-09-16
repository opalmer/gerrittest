package gerrittest

import (
	"testing"

	. "gopkg.in/check.v1"
)

type ChangeTest struct{}

var _ = Suite(&ChangeTest{})

func (s *ChangeTest) TestChange(c *C) {
	if testing.Short() {
		c.Skip("-short provided")
	}

	cfg := NewConfig()
	gerrit, err := New(cfg)
	c.Assert(err, IsNil)
	defer gerrit.Destroy() // nolint: errcheck

	// FIXME: Commit should not be required first
	c.Assert(gerrit.Repo.Commit("testing"), IsNil)
	change, err := gerrit.CreateChange("foobar")
	c.Assert(err, IsNil)

	c.Assert(change.Write("foo/bar.txt", 0600, []byte("hello")), IsNil)
	c.Assert(change.AmendAndPush(), IsNil)
	c.Assert(change.Remove("foo"), IsNil)
	c.Assert(change.AmendAndPush(), IsNil)

	// TODO This is failing for some reason
	//_, err = change.AddFileComment("", "foo/bar.txt", 1, "Test comment.")
	//c.Assert(err, IsNil)

	_, err = change.ApplyLabel("", "Code-Review", 2)
	c.Assert(err, IsNil)

	_, err = change.ApplyLabel("", "Verified", 1)
	c.Assert(err, IsNil)

	_, err = change.AddTopLevelComment("", "looks good")
	c.Assert(err, IsNil)

	_, err = change.Submit()
	c.Assert(err, IsNil)

}
