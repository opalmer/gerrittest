package gerrittest

import (
	"flag"
	"os"
	"testing"

	"github.com/opalmer/logrusutil"
	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

var (
	testLogLevel = flag.String(
		"gerrittest.log-level", "panic",
		"Controls the log level for the logging package.")
)

func Test(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *testLogLevel != "" {
		cfg := logrusutil.NewConfig()
		cfg.Level = *testLogLevel
		if err := logrusutil.ConfigureLogger(log.StandardLogger(), cfg); err != nil {
			log.WithError(err).Panic()
			os.Exit(1)
		}
	}

	TestingT(t)
}
