package gerrittest

import (
	"testing"

	. "gopkg.in/check.v1"
)

type GerritTest struct{}

var _ = Suite(&GerritTest{})

func (s *GerritTest) TestNew(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}
	cfg := NewConfig()
	gerrit, err := New(cfg)
	c.Assert(err, IsNil)
	c.Assert(gerrit.Destroy(), IsNil)
}
