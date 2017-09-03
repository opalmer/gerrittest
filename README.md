# Gerrit Testing With Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)
[![codecov](https://codecov.io/gh/opalmer/gerrittest/branch/master/graph/badge.svg)](https://codecov.io/gh/opalmer/gerrittest)
[![Go Report Card](https://goreportcard.com/badge/github.com/opalmer/gerrittest)](https://goreportcard.com/report/github.com/opalmer/gerrittest)
[![GoDoc](https://godoc.org/github.com/opalmer/gerrittest?status.svg)](https://godoc.org/github.com/opalmer/gerrittest)

This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers. Documentation 
is available via godoc: https://godoc.org/github.com/opalmer/gerrittest

## Setup

* Install docker
* `go install github.com/opalmer/gerrittest/cmd`

## Command Line Usage
### Start

```
$ gerrittest start
{
  "config": {
    "image": "opalmer/gerrittest:2.14.3",
    "port_ssh": 0,
    "port_http": 0,
    "repo_root": "",
    "private_key": "",
    "username": "admin",
    "password": ""
  },
  "container": {
    "http": {
      "Private": 8080,
      "port": 37573,
      "address": "localhost",
      "protocol": "tcp"
    },
    "ssh": {
      "Private": 29418,
      "port": 32787,
      "address": "127.0.0.1",
      "protocol": "tcp"
    },
    "image": "opalmer/gerrittest:2.14.3",
    "id": "25482db97051b0317a14e8271c36947e610a56c18550b405aa0d441c09e7947a"
  },
  "http": {
    "Private": 8080,
    "port": 37573,
    "address": "localhost",
    "protocol": "tcp"
  },
  "ssh": {
    "Private": 29418,
    "port": 32787,
    "address": "127.0.0.1",
    "protocol": "tcp"
  },
  "repo": {
    "path": "/tmp/gerrittest-093084676"
  },
  "private_key_path": "/tmp/gerrittest-id_rsa-186419449",
  "username": "admin",
  "password": "l7aJMAr70ThKMTame0ZEZr/cFH4pJnrasEaNEadlTQ"
}
```

## Code Examples

Visit godoc.org to see code examples:

https://godoc.org/github.com/opalmer/gerrittest#pkg-examples

## Testing

The gerrittest project can be tested locally. To build the container and
the gerrittest command run:

```
$ make check
```

You can also skip some of the slower tests:

```
$ go test -v -short github.com/opalmer/gerrittest
```

If you're having trouble with a specific test you can enable debug 
logging and run that test specifically:

```
$ go test -gerrittest.loglevel=debug -check.vv -check.f RepoTest.* github.com/opalmer/gerrittest
```
