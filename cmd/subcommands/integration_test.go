package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/opalmer/gerrittest"
	. "gopkg.in/check.v1"
)

type IntegrationTest struct{}

var _ = Suite(&IntegrationTest{})

func (s *IntegrationTest) Test_StartStop(c *C) {
	if testing.Short() {
		c.Skip("-short set")
	}
	privateFile, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	defer os.Remove(privateFile.Name()) // nolint: errcheck

	// Generate a custom private key and pass it to Start(). This is
	// important to do because we want to replicate a case where a user
	// provides their own key and wants to cleanup after a test run. The
	// expectation is that gerrittest *does not* delete the key during
	// cleanup.
	private, err := gerrittest.GenerateRSAKey()
	c.Assert(err, IsNil)
	c.Assert(gerrittest.WriteRSAKey(private, privateFile), IsNil)

	file, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	defer os.Remove(file.Name()) // nolint: errcheck
	flags := []string{"--json", file.Name(), "--private-key", privateFile.Name()}

	// Call the start subcommand and write the struct to disk.
	c.Assert(file.Close(), IsNil)
	c.Assert(Start.ParseFlags(flags), IsNil)
	c.Assert(Start.RunE(Start, []string{}), IsNil)

	// Load information from json and then perform a few common
	// operations.
	gerrit, err := gerrittest.NewFromJSON(file.Name())
	c.Assert(err, IsNil)
	defer gerrit.Container.Terminate()   // nolint: errcheck
	defer os.RemoveAll(gerrit.Repo.Path) // nolint: errcheck

	client, err := gerrit.HTTP.Gerrit()
	c.Assert(err, IsNil)

	_, _, err = client.Projects.CreateProject("testing", nil)
	c.Assert(err, IsNil)

	// Add a remote based on the container then push a commit to it.
	c.Assert(gerrit.Repo.AddRemoteFromContainer(gerrit.Container, "", "testing"), IsNil)
	c.Assert(gerrit.Repo.AddContent("foo/bar.txt", 0600, []byte("Hello")), IsNil)
	c.Assert(gerrit.Repo.Commit("42: hello"), IsNil) // Note, the number is an attempt to mess up Push()
	c.Assert(gerrit.Repo.Push("", ""), IsNil)
	change, err := gerrit.Repo.ChangeID()
	c.Assert(err, IsNil)

	// Use the change id retrieved from git to retrieve information
	// about the change.
	_, _, err = client.Changes.GetChangeDetail(change, nil)
	c.Assert(err, IsNil)

	// Read the struct from disk and use it to stop the container. This should
	// also cleanup the repository.
	c.Assert(Stop.ParseFlags([]string{"--json", file.Name()}), IsNil)
	c.Assert(Stop.RunE(Stop, []string{}), IsNil)

	// The below test to make sure that the custom private key and
	// repository were not removed when cleanup was run.
	_, err = os.Stat(privateFile.Name())
	c.Assert(err, IsNil)
	_, err = os.Stat(gerrit.Repo.Path)
	c.Assert(os.IsNotExist(err), Equals, true)
}
