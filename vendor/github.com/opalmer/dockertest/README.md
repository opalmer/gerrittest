# Docker Test

[![Build Status](https://travis-ci.org/opalmer/dockertest.svg?branch=master)](https://travis-ci.org/opalmer/dockertest)
[![codecov](https://codecov.io/gh/opalmer/dockertest/branch/master/graph/badge.svg)](https://codecov.io/gh/opalmer/dockertest)
[![Go Report Card](https://goreportcard.com/badge/github.com/opalmer/dockertest)](https://goreportcard.com/report/github.com/opalmer/dockertest)
[![GoDoc](https://godoc.org/github.com/opalmer/dockertest?status.svg)](https://godoc.org/github.com/opalmer/dockertest)

This project provides a small set of wrappers around docker. It is intended
to be used to ease testing. Documentation is available via godoc: 
    https://godoc.org/github.com/opalmer/dockertest

# Examples

Create a container and retrieve an exposed port.

```go
import (
	"context"
	"log"
	"github.com/opalmer/dockertest"
)

func main() {
	client, err := dockertest.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Construct information about the container to start.
	input := dockertest.NewClientInput("nginx:mainline-alpine")
	input.Ports.Add(&dockertest.Port{
		Private:  80,
		Public:   dockertest.RandomPort,
		Protocol: dockertest.ProtocolTCP,
	})

	// Start the container
	container, err := client.RunContainer(context.Background(), input)
	if err != nil {
		log.Fatal(err)
	}

	// Extract information about the started container.
	port, err := container.Port(80)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(port.Public, port.Address)

	if err := client.RemoveContainer(context.Background(), container.ID()); err != nil {
		log.Fatal(err)
	}
}
```

Create a container using the `Service` struct.

```go
import (
	"context"
	"log"
	"github.com/opalmer/dockertest"
)

func main() {
	client, err := dockertest.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Construct information about the container to start.
	input := dockertest.NewClientInput("nginx:mainline-alpine")
	input.Ports.Add(&dockertest.Port{
		Private:  80,
		Public:   dockertest.RandomPort,
		Protocol: dockertest.ProtocolTCP,
	})

	// Construct the service and tell it how to handle waiting
	// for the container to start.
	service := client.Service(input)
	service.Ping = func(input *dockertest.PingInput) error {
		port, err := input.Container.Port(80)
		if err != nil {
			return err // Will cause Run() to call Terminate()
		}

		for {
			_, err := net.Dial(string(port.Protocol), fmt.Sprintf("%s:%d", port.Address, port.Public))
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			break
		}

		return nil
	}

	// Starts the container, runs Ping() and waits for it to return. If Ping()
	// fails the container will be terminated and Run() will return an error.
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}

	// Container has started, get information information
	// about the exposed port.
	port, err := service.Container.Port(80)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(port.Public, port.Address)

	if err := service.Terminate(); err != nil {
		log.Fatal(err)
	}
}
```
