package gerrittest

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/opalmer/dockertest"
)

// You can start the Gerrit container using the NewContainer() function. In
// this example an random port will be used for both http and the default
// image will be used. This kind of setup is useful if you don't want to
// gerrittest to perform any setup steps for you.
func ExampleNewContainer() {
	container, err := NewContainer(
		context.Background(), dockertest.RandomPort, dockertest.RandomPort, "")
	if err != nil {
		log.Fatal(err)
	}

	// Terminate the container when you're done.
	if err := container.Terminate(); err != nil {
		log.Fatal(err)
	}
}

// Once you've started the service you'll want to setup Gerrit inside
// the container. Running Setup.Init will cause the administrative user to
// be created, generate an http API password and insert a public key for ssh
// access.
func ExampleNew() {
	cfg := NewConfig()
	gerrit, err := New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	files := map[string]string{
		"README.md":        "# Hello",
		"scripts/foo.bash": "echo 'foo'",
	}
	for relative, content := range files {
		path := filepath.Join(gerrit.Config.RepoRoot, relative)
		if err := ioutil.WriteFile(path, []byte(content), 0600); err != nil {
			log.Fatal(err)
		}
		if err := gerrit.Repo.Add(relative); err != nil {
			log.Fatal(err)
		}
	}

	if err := gerrit.Repo.Commit("test: first change"); err != nil {
		log.Fatal(err)
	}

	if err := gerrit.Repo.Push("origin", "refs/for/master:HEAD"); err != nil {
		log.Fatal(err)
	}

	// Terminate the container and cleanup all data related to
	// this run.
	if err := gerrit.Destroy(); err != nil {
		log.Fatal(err)
	}
}
