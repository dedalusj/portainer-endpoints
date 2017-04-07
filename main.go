package main

import (
	"os"
	"github.com/urfave/cli"

	"github.com/Sirupsen/logrus"
)

var version string

func createLogger(c *cli.Context) *logrus.Logger {
	logger := logrus.New()
	logger.Out = os.Stderr
	logger.Formatter = &logrus.TextFormatter{}
	logger.Level = logrus.InfoLevel
	if c.Bool("debug") {
		logger.Level = logrus.DebugLevel
	}
	return logger
}

func fetchEndpoints(c *cli.Context) {
	logger := createLogger(c)

	logger.WithFields(logrus.Fields{
		"version": version,
	}).Info("Portainer endpoints manager")

}

func main() {
	app := cli.NewApp()
	app.Name = "go-deploy"
	app.Usage = "Command line tool for blue/green deployment of docker containers"
	app.Version = version
	app.Action = fetchEndpoints
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Enable debug logging",
			EnvVar: "DEBUG",
		},
	}

	app.Run(os.Args)
}
