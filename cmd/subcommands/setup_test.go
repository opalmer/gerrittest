package cmd

import (
	"flag"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

var (
	testLogLevel = flag.String(
		"gerrittest.loglevel", "panic",
		"Controls the log level for the logging package.")
)

func Test(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}
	if *testLogLevel != "" {
		level, err := log.ParseLevel(*testLogLevel)
		if err != nil {
			os.Exit(1)
		}
		log.SetLevel(level)
	}

	TestingT(t)
}
