package main

import (
	"github.com/opalmer/gerrittest"
	"github.com/opalmer/logrusutil"
	log "github.com/sirupsen/logrus"
)

func chkerr(err error) {
	if err != nil {
		panic(err)
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

	chkerr(change.Write("foo/bar", 0600, []byte("hello")))
	chkerr(change.AmendAndPush())
	chkerr(change.Remove("foo"))
	chkerr(change.AmendAndPush())
	_, err = change.ApplyLabel("", gerrittest.CodeReviewLabel, 2)
	chkerr(err)
	_, err = change.AddTopLevelComment("", "Hello, world")
	chkerr(err)
	_, err = change.ApplyLabel("", gerrittest.VerifiedLabel, 1)
	chkerr(err)

}
