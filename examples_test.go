package gerrittest

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
)

// You can start the Gerrit service using the Start() function. This only
// starts the container and returns information about the service. Useful if
// you only need the service and don't need or want a repository, http client,
// initial admin user, etc.
func ExampleStart() {
	svc, err := Start(context.Background(), NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Terminate the container when you're done.
	if err := svc.Service.Terminate(); err != nil {
		log.Fatal(err)
	}
}

// Once you've started the service you'll want to setup Gerrit inside
// the container. Running Setup.Init will cause the administrative user to
// be created, generate an http api password and insert a public key for ssh
// access.
func ExampleSetup() {
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
		path := filepath.Join(gerrit.Repo.Path, relative)
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
