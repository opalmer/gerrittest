package gerrittest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	. "gopkg.in/check.v1"
)

type JSONTest struct{}

var _ = Suite(&ConfigTest{})

func newSpec(c *C) (*ServiceSpec, string) {
	spec := &ServiceSpec{Container: "foo"}
	data, err := json.Marshal(spec)
	c.Assert(err, IsNil)

	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)

	_, err = io.Copy(file, bytes.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)

	return spec, file.Name()
}

func (s *JSONTest) TestReadServiceSpec(c *C) {
	spec, path := newSpec(c)
	specFromFile, err := ReadServiceSpec(path)
	c.Assert(err, IsNil)
	c.Assert(specFromFile.Container, Equals, spec.Container)
	c.Assert(os.Remove(path), IsNil)
}

func (s *JSONTest) TestReadServiceSpecFromArg(c *C) {
	spec, path := newSpec(c)
	defer os.Remove(path)

	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "")
	c.Assert(cmd.Flags().Set("json", path), IsNil)
	specFromFile, err := ReadServiceSpecFromArg(cmd)
	c.Assert(err, IsNil)
	c.Assert(specFromFile.Container, Equals, spec.Container)
	c.Assert(os.Remove(path), IsNil)
}
