package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/opalmer/gerrittest"
	"github.com/opalmer/logrusutil"
)

func chkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log.SetLevel(log.DebugLevel)
	logcfg := logrusutil.NewConfig()
	logcfg.Level = "debug"
	chkerr(logrusutil.ConfigureLogger(log.StandardLogger(), logcfg))

	gerrit, err := gerrittest.NewFromJSON("/tmp/gerrit.json")
	chkerr(err)
	change, err := gerrit.CreateChange("foobar")

	chkerr(err)
	_ = change

	//cfg := gerrittest.NewConfig()
	//gerrit, err := gerrittest.New(cfg)

	//chkerr(err)
	//defer gerrit.Destroy()
	//log.Warn("!!! SUCCESS !!!")
}
