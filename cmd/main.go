package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
)

var (
	image = flag.String(
		"image", "opalmer/gerrittest:latest",
		"The Docker image to use to run Gerrit.")
	portHTTP = flag.Uint(
		"http", uint(dockertest.RandomPort),
		"The port to map to the HTTP service. Random by default.")
	portSSH = flag.Uint(
		"ssh", uint(dockertest.RandomPort),
		"The port to map to the HTTP service. Random by default.")
	debug = flag.Bool(
		"debug", false,
		"If provided enable debug logging")
)

func main() {
	flag.Parse()
	client, err := dockertest.NewClient()
	if err != nil {
		panic(err)
	}

	cfg := gerrittest.NewConfig()
	cfg.Image = *image
	cfg.PortSSH = uint16(*portSSH)
	cfg.PortHTTP = uint16(*portHTTP)
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	service := gerrittest.NewService(client, cfg)
	if err := service.Run(); err != nil {
		panic(err)
	}
}
