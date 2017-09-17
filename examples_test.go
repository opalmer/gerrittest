package gerrittest

import (
	"context"
	"log"

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
func ExampleGerrit_CreateChange() {
	cfg := NewConfig()
	gerrit, err := New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	files := map[string]string{
		"README.md":        "# Hello",
		"scripts/foo.bash": "echo 'foo'",
	}

	change, err := gerrit.CreateChange("testing", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer change.Destroy() // nolint: errcheck
	defer gerrit.Destroy() // nolint: errcheck

	for relative, content := range files {
		if err := change.Add(relative, 0600, content); err != nil {
			log.Fatal(err)
		}
	}
	if err := change.Push(); err != nil {
		log.Fatal(err)
	}
	if _, err := change.ApplyLabel("1", CodeReviewLabel, 2); err != nil {
		log.Fatal(err)
	}
	if _, err := change.ApplyLabel("1", VerifiedLabel, 1); err != nil {
		log.Fatal(err)
	}
	if _, err := change.Submit(); err != nil {
		log.Fatal(err)
	}

}
