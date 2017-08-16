package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
)

func TestStart(t *testing.T) {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	os.Remove(file.Name())
	defer os.Remove(file.Name())

	if err := Start.Flags().Parse([]string{"--json", file.Name()}); err != nil {
		t.Fatal(err)
	}

	if err := Start.RunE(Start, []string{}); err != nil {
		t.Fatal(err)
	}

	output, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	spec := &gerrittest.ServiceSpec{}
	if err := json.Unmarshal(output, spec); err != nil {
		t.Fatal(err)
	}

	client, err := dockertest.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveContainer(context.Background(), spec.Container); err != nil {
		t.Fatal(err)
	}
}
