package gerrittest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func newSpec(t *testing.T) (*ServiceSpec, string) {
	spec := &ServiceSpec{Container: "foo"}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}

	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(file, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return spec, file.Name()
}

func TestReadServiceSpec(t *testing.T) {
	spec, path := newSpec(t)
	defer os.Remove(path)

	specFromFile, err := ReadServiceSpec(path)
	if err != nil {
		t.Fatal(err)
	}
	if specFromFile.Container != spec.Container {
		t.Fatal(err)
	}
}

func TestReadServiceSpecFromArg(t *testing.T) {
	spec, path := newSpec(t)
	defer os.Remove(path)

	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "")

	if err := cmd.Flags().Set("json", path); err != nil {
		t.Fatal(err)
	}
	specFromFile, err := ReadServiceSpecFromArg(cmd)
	if err != nil {
		t.Fatal(err)
	}

	if specFromFile.Container != spec.Container {
		t.Fatal(err)
	}
}
