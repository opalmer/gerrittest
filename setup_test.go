package gerrittest

import (
	"flag"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
)

var (
	logLevelName = flag.String(
		"gerrittest.loglevel", "panic",
		"Controls the log level for the logging package.")
)

func TestMain(m *testing.M) {
	flag.Parse()

	if *logLevelName != "" {
		level, err := log.ParseLevel(*logLevelName)
		if err != nil {
			os.Exit(1)
		}
		log.SetLevel(level)
	}

	os.Exit(m.Run())
}
