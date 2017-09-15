package main

import (
	"github.com/opalmer/gerrittest"
	"github.com/opalmer/logrusutil"
	log "github.com/sirupsen/logrus"
)

const codePath = "scripts/example.sh"
const code = `
#!/bin/bash

echo "Hello, world"
`

func chkerr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	logcfg := logrusutil.NewConfig()
	logcfg.Level = "debug"
	chkerr(logrusutil.ConfigureLogger(log.StandardLogger(), logcfg))

	gerrit, err := gerrittest.NewFromJSON("/tmp/gerrit.json")
	chkerr(err)
	_ = gerrit
	change, err := gerrit.CreateChange("foobar")
	chkerr(err)

	chkerr(change.Write(codePath, 0600, []byte(code)))
	chkerr(change.AmendAndPush())
	chkerr(change.Remove("foo"))
	chkerr(change.AmendAndPush())

	_, err = change.AddFileComment("2", codePath, 2, "Test comment.")
	chkerr(err)

	_, err = change.ApplyLabel("", gerrittest.CodeReviewLabel, 2)
	chkerr(err)
	_, err = change.AddTopLevelComment("", "Looks good!")
	chkerr(err)
	_, err = change.ApplyLabel("", gerrittest.VerifiedLabel, 1)
	chkerr(err)

}
