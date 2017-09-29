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
### Start and Stop

```
$ go get github.com/opalmer/gerrittest
$ cd ~/go/src/github.com/opalmer/gerrittest
$ make dep build
$ ./gerrittest start --json /tmp/gerrit.json
$ cat /tmp/gerrit.json
{
  "config": {
    "image": "opalmer/gerrittest:2.14.3",
    "port_ssh": 0,
    "port_http": 0,
    "timeout": 300000000000,
    "git": {
      "core.sshCommand": "ssh -i /tmp/gerrittest-id_rsa-706055562 -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no",
      "user.email": "admin@localhost",
      "user.name": "admin"
    },
    "ssh_keys": [
      {
        "path": "/tmp/gerrittest-id_rsa-706055562",
        "generated": true,
        "default": true
      }
    ],
    "username": "admin",
    "password": "oD7BNb6YE21+7ZGEXefJtFk3HY85wKYrfiZg13H6Mg",
    "skip_setup": false,
    "cleanup_container": true
  },
  "container": {
    "http": {
      "Private": 8080,
      "port": 33511,
      "address": "localhost",
      "protocol": "tcp"
    },
    "ssh": {
      "Private": 29418,
      "port": 32791,
      "address": "127.0.0.1",
      "protocol": "tcp"
    },
    "image": "opalmer/gerrittest:2.14.3",
    "id": "6ef42639c9a40aa3a5e793b8d7fe33005e585ae1ce636671e1bb2d15fc8b1173"
  },
  "http": {
    "Private": 8080,
    "port": 33511,
    "address": "localhost",
    "protocol": "tcp"
  },
  "ssh": {
    "Private": 29418,
    "port": 32791,
    "address": "127.0.0.1",
    "protocol": "tcp"
  }
}
$ ./gerrittest stop --json /tmp/gerrit.json
```

### Combining gerrittest, bash and curl

```bash

$ JSON="/tmp/services.json"
$ PREFIX=")]}'"
$ gerrittest start --json "$JSON"
$ USERNAME="$(jq -r ".username" "$JSON")"
$ PASSWORD="$(jq -r ".password" "$JSON")"
$ URL="http://$(jq -r ".http.address" "$JSON"):$(jq -r ".http.port" "$JSON")"
$ RAW_RESPONSE="$(curl -u $USERNAME:$PASSWORD $URL/a/accounts/self --fail --silent)"
$ RESPONSE=$(echo "$RAW_RESPONSE" | sed -e "s/^$PREFIX//")
$ echo "$RESPONSE" | jq ._account_id
1000000
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
$ go test -gerrittest.log-level=debug -check.vv -check.f RepoTest.* github.com/opalmer/gerrittest
```
