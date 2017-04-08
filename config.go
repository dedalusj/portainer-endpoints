package main

import (
	"time"
	"github.com/urfave/cli"
	"fmt"
)

type DockerConfig struct {
	Port   int
	TLS    bool
	CACert string
	Cert   string
	Key    string
}

type Config struct {
	Debug      bool
	Tag        string
	Interval   time.Duration
	OutputFile string
	Docker     DockerConfig
}

var defaultConfig = &Config{
	Tag: "stack_name",
	Interval: time.Minute,
	Docker: DockerConfig{
		Port: 2376,
	},
}

func validate(docker *DockerConfig) error {
	if docker.CACert == "" && docker.Cert == "" && docker.Key == "" {
		return nil
	}

	if docker.CACert != "" && docker.Cert != "" && docker.Key != "" {
		docker.TLS = true
		return nil
	}

	return fmt.Errorf("All three TLS options must be specified. Found ca-cert=%s cert=%s key=%s",
		docker.CACert, docker.Cert, docker.Key)
}

func NewConfig(c *cli.Context) (*Config, error) {
	config := defaultConfig

	config.Debug = c.Bool("debug")
	if t := c.String("tag"); t != "" {
		config.Tag = t
	}
	if i := c.Duration("interval"); i != 0 {
		config.Interval = i
	}
	config.OutputFile = c.String("output-file")

	if p := c.Int("docker-port"); p != 0 {
		config.Docker.Port = p
	}
	config.Docker.CACert = c.String("ca-cert")
	config.Docker.Cert = c.String("cert")
	config.Docker.Key = c.String("key")
	err := validate(&config.Docker)
	if err != nil {
		return config, err
	}

	return config, nil
}
