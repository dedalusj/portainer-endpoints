package main  // import "github.com/dedalusj/portainer-endpoints"

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
)

var version string

func createLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Out = os.Stderr
	logger.Formatter = &logrus.TextFormatter{}
	logger.Level = logrus.InfoLevel
	return logger
}

func writeToFile(filepath, content string) error {
	fo, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = fmt.Fprint(fo, content)
	return err
}

func fetchEndpoints(config *Config) error {
	instances, err:= getFilteredInstances(config.Tag)
	if err != nil {
		return err
	}

	endpoints := createEndpoints(instances, config.Docker)
	b, err := json.Marshal(endpoints)
	if err != nil {
		return err
	}

	if config.OutputFile == "" {
		fmt.Print(string(b))
	} else {
		writeToFile(config.OutputFile, string(b))
	}

	return nil
}

func run(c *cli.Context) {
	logger := createLogger()
	config, err := NewConfig(c)
	if err != nil {
		logger.Fatal(err)

	}

	logger.WithField("version", version).Info("Portainer endpoints manager")

	for {
		time.Sleep(config.Interval)
		err := fetchEndpoints(config)
		if err != nil {
			logger.WithField("tag", config.Tag).Warnf("Error while fetching endpoints: %+v", err)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "portainer-endpoints"
	app.Usage = "Command line tool for generating a json file with Docker endpoints for Portainer"
	app.Version = version
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Enable debug logging",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "tag, t",
			Usage:  "Tag used to filter EC2 instances",
			EnvVar: "TAG",
		},
		cli.IntFlag{
			Name:   "port, p",
			Usage:  "Port for the Docker remote API",
			EnvVar: "PORT",
		},
		cli.DurationFlag{
			Name:   "interval, i",
			Usage:  "Interval for querying the EC2 isntances",
			EnvVar: "INTERVAL",
		},
		cli.StringFlag{
			Name:   "ca-cert",
			Usage:  "Path to the ca certificate",
			EnvVar: "CA_CERT",
		},
		cli.StringFlag{
			Name:   "cert",
			Usage:  "Path to the certificate",
			EnvVar: "CERT",
		},
		cli.StringFlag{
			Name:   "key",
			Usage:  "Path to the key",
			EnvVar: "KEY",
		},
		cli.StringFlag{
			Name:   "output-file, o",
			Usage:  "Path of the output json file. If unspecified it will write to stdout",
			EnvVar: "OUTPUT_FILE",
		},
	}

	app.Run(os.Args)
}
