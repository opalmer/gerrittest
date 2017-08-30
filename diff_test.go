package gerrittest

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "gopkg.in/check.v1"
)

const patch = `
From 4460c2d5b1d762394f9d2e28e7d00e0dbf6ba7ff Mon Sep 17 00:00:00 2001
From: Oliver Palmer <oliverpalmer@opalmer.com>
Date: Sat, 26 Aug 2017 15:06:18 -0400
Subject: [PATCH] hello, world

---
 hello_world.txt | 1 +
 1 file changed, 1 insertion(+)
 create mode 100644 hello_world.txt

diff --git a/hello_world.txt b/hello_world.txt
new file mode 100644
index 0000000..f75ba05
--- /dev/null
+++ b/hello_world.txt
@@ -0,0 +1 @@
+Hello, world.
--
2.13.5
`

type DiffTest struct{}

var _ = Suite(&DiffTest{})

func (s *DiffTest) TestApplyToRoot_Error(c *C) {
	expected := errors.New("Testing")
	diff := &Diff{Error: expected}
	c.Assert(diff.ApplyToRoot(""), ErrorMatches, expected.Error())
}

func (s *DiffTest) TestApplyToRoot(c *C) {
	diff := &Diff{Content: []byte(patch)}
	path, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	cmd := exec.Command("git", "init", path)
	c.Assert(cmd.Run(), IsNil)
	c.Assert(diff.ApplyToRoot(path), IsNil)
	content, err := ioutil.ReadFile(filepath.Join(path, "hello_world.txt"))
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, []byte("Hello, world.\n"))
	c.Assert(os.RemoveAll(path), IsNil)
}
