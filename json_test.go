package gerrittest

import (
	. "gopkg.in/check.v1"
	"encoding/json"
	"io/ioutil"
	"os"
	"io"
	"bytes"
)

type JSONTest struct{}

var _ = Suite(&JSONTest{})

func (s *JSONTest) TestReadServiceSpec(c *C) {
	spec := ServiceSpec{Container: "foo"}
	data, err := json.Marshal(spec)
	c.Assert(err, IsNil)
	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	defer os.Remove(file.Name())
	_, err = io.Copy(file, bytes.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(file.Close(), IsNil)
	specFromFile, err := ReadServiceSpec(file.Name())
	c.Assert(err, IsNil)
	c.Assert(specFromFile.Container, DeepEquals, "foo")
}