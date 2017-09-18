# Logrus Wrapper

[![Build Status](https://travis-ci.org/opalmer/logrusutil.svg?branch=master)](https://travis-ci.org/opalmer/logrusutil)
[![codecov](https://codecov.io/gh/opalmer/logrusutil/branch/master/graph/badge.svg)](https://codecov.io/gh/opalmer/logrusutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/opalmer/logrusutil)](https://goreportcard.com/report/github.com/opalmer/logrusutil)
[![GoDoc](https://godoc.org/github.com/opalmer/logrusutil?status.svg)](https://godoc.org/github.com/opalmer/logrusutil)

This project provides very basic wrappers for [logrus](https://github.com/sirupsen/logrus)
including:

* Log level parsing and handling.
* Hooks for adding additional information to log messages.

Contributions are welcome but before opening a PR consider if your request would
be better served as a contribution directly to the logrus project. This project
was initially created so a few different projects could share the same code.

```go
package main

import (
	"github.com/opalmer/logrusutil"
	"github.com/Sirupsen/logrus"
)

func main() {
	// Setup the root logger and hooks.
	if err := logrusutil.ConfigureLogger(logrus.StandardLogger(), logrusutil.NewConfig()); err != nil {
		panic(err)
	}
}
```

## Case Sensitivity

The logrus project was renamed at one point from:

    github.com/Sirupsen/logrus

To:

    github.com/sirupsen/logrus

This causes conflicts to occur in certain cases if one or more of your
dependencies are using the old import path. This project will `Sirupsen` for
the time being but reserves the right to change this in the future.
