# Gerrit Testing With Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)
[![codecov](https://codecov.io/gh/opalmer/gerrittest/branch/master/graph/badge.svg)](https://codecov.io/gh/opalmer/gerrittest)
[![Go Report Card](https://goreportcard.com/badge/github.com/opalmer/gerrittest)](https://goreportcard.com/report/github.com/opalmer/gerrittest)
[![GoDoc](https://godoc.org/github.com/opalmer/dockertest?status.svg)](https://godoc.org/github.com/opalmer/dockertest)

This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers. Documentation 
is available via godoc: https://godoc.org/github.com/opalmer/gerrittest

## Setup

* Install docker
* `go install github.com/opalmer/gerrittest/cmd`

## Usage

### Command Line
#### gerrittest - start

```
$ gerrittest start
{
 "admin": {
  "login": "admin",
  "password": "+YzOzJ9xBftJnvyWrSOSHqrviFlPCP2J7IPxUspKNg",
  "private_key": "/tmp/id_rsa-158272732"
 },
 "container": "b90671cb7d192131102cd599df5cfa4d4b4ca78f6857da0a41272f2063a22530",
 "ssh": {
  "Private": 29418,
  "port": 32783,
  "address": "127.0.0.1",
  "protocol": "tcp"
 },
 "http": {
  "Private": 8080,
  "port": 36965,
  "address": "127.0.0.1",
  "protocol": "tcp"
 },
 "url": "http://127.0.0.1:36965",
 "ssh_command": "ssh -p 32783 -i /tmp/id_rsa-158272732 -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no admin@127.0.0.1"
}
```

### API
### Basic Usage

The produces nearly almost identical results to `gerrittest start` above.

```go
import (
	"context"
	"github.com/opalmer/gerrittest"
)

func main()  {
	service, err := gerrittest.Start(context.Background(), gerrittest.NewConfig())
	if err != nil {
		panic(err)
	}
	setup := &Setup{Service: service}
	spec, httpClient, sshClient, err := setup.Init()
	if err != nil {
		panic(err)
	}
}
```


## Testing

The gerrittest project can be tested locally. To build the container and
the gerrittest command run:

```
$ make check
```
