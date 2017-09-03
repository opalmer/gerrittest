package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	. "gopkg.in/check.v1"
)

type UtilTest struct{}

var _ = Suite(&UtilTest{})

func (s *UtilTest) SetUpTest(c *C) {
	exit = false
}

func (s *UtilTest) TearDownTest(c *C) {
	exit = true
}

func (s *UtilTest) TestExitIf(c *C) {
	c.Assert(exitIf("", nil), Equals, false)
	c.Assert(exitIf("", errors.New("hello")), Equals, true)
}

func (s *UtilTest) TestGetBool(c *C) {
	command := &cobra.Command{}
	command.Flags().Bool("test", false, "")
	c.Assert(command.ParseFlags([]string{"--test"}), IsNil)
	c.Assert(getBool(command, "test"), Equals, true)
}

func (s *UtilTest) TestGetString(c *C) {
	command := &cobra.Command{}
	command.Flags().String("test", "", "")
	c.Assert(command.ParseFlags([]string{"--test=bar"}), IsNil)
	c.Assert(getString(command, "test"), Equals, "bar")
}

func (s *UtilTest) TestGetDuration(c *C) {
	command := &cobra.Command{}
	command.Flags().Duration("test", time.Second*0, "")
	c.Assert(command.ParseFlags([]string{"--test=15m"}), IsNil)
	c.Assert(getDuration(command, "test"), Equals, time.Minute*15)
}

func (s *UtilTest) TestGetUInt16(c *C) {
	command := &cobra.Command{}
	command.Flags().Uint16("test", 0, "")
	c.Assert(command.ParseFlags([]string{"--test=65535"}), IsNil)
	c.Assert(getUInt16(command, "test"), Equals, uint16(65535))
}

func (s *UtilTest) Test_jsonOutput_stdout(c *C) {
	old := os.Stdout // keep backup of the real stdout
	r, w, err := os.Pipe()
	c.Assert(err, IsNil)
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	output := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) // nolint: errcheck
		output <- buf.String()
	}()

	command := &cobra.Command{}
	command.Flags().String("json", "", "")
	c.Assert(jsonOutput(command, &gerrittest.Gerrit{Username: "foobar"}), IsNil)
	w.Close() // nolint: errcheck
	c.Assert(strings.Contains(<-output, "foobar"), Equals, true)
}

func (s *UtilTest) Test_jsonOutput_file(c *C) {
	command := &cobra.Command{}
	command.Flags().String("json", "", "")
	file, err := ioutil.TempFile("", "")
	defer os.Remove(file.Name()) // nolint: errcheck
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	c.Assert(command.ParseFlags(
		[]string{fmt.Sprintf("--json=%s", file.Name())}), IsNil)
	c.Assert(jsonOutput(command, &gerrittest.Gerrit{Username: "foobar"}), IsNil)
	data, err := ioutil.ReadFile(file.Name())
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(string(data), "foobar"), Equals, true)
}
