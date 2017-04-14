package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	envPrefix = "PE_"
)

var version string

type Config struct {
	Tag      string
	Output   string
	Port     int
	Interval time.Duration
	Debug    bool
}

type Endpoint struct {
	Name string
	URL  string
}

type Instance struct {
	Name string
	Ip   string
}

func NewInstance(instance *ec2.Instance) Instance {
	ip := aws.StringValue(instance.PrivateIpAddress)
	name := strings.Replace(ip, ".", "-", -1)
	for _, t := range instance.Tags {
		if strings.ToLower(aws.StringValue(t.Key)) == "name" {
			name = strings.ToLower(aws.StringValue(t.Value)) + "-" + name
		}
	}
	return Instance{Name: name, Ip: ip}
}

func (i Instance) GetEndpoint(port int) Endpoint {
	url := fmt.Sprintf("tcp://%s:%d", i.Ip, port)
	return Endpoint{
		Name: i.Name,
		URL: url,
	}
}

type Tag struct {
	Key   string
	Value string
}

func NewTag(tag string) (Tag, error) {
	pieces := strings.SplitN(tag, "=", 2)
	if len(pieces) < 2 {
		return Tag{}, fmt.Errorf("Invalid tag [%s]. Expected tag=value format", tag)
	}
	t := Tag{Key: pieces[0], Value: pieces[1]}
	log.WithFields(log.Fields{
		"name": t.Key,
		"value": t.Value,
	}).Info("Parsed tag")
	return t, nil
}

func (t Tag) String() string {
	return fmt.Sprintf("%s=%s", t.Key, t.Value)
}

type Clients struct {
	EC2Client     ec2iface.EC2API
}

func NewEC2Client() ec2iface.EC2API {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))},
	}))
	return ec2.New(s)
}

func initLogging(debug bool) {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func getInstances(tag Tag, client ec2iface.EC2API) ([]Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + tag.Key),
				Values: aws.StringSlice([]string{tag.Value}),
			},
			{
				Name: aws.String("instance-state-name"),
				Values: aws.StringSlice([]string{"pending", "running"}),
			},
		},
	}

	resp, err := client.DescribeInstances(params)
	if err != nil {
		return []Instance{}, errors.Wrapf(err, "Describing instances with tag [%s]", tag)
	}

	instances := []Instance{}
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, NewInstance(i))
		}
	}

	log.WithFields(log.Fields{
		"num": len(instances),
		"tag": tag,
	}).Debug("Fetched instances")
	return instances, nil
}

func writeEndpoints(endpoints []Endpoint, output string) error {
	b, err := json.Marshal(endpoints)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal endpoints")
	}

	if output == "" {
		fmt.Printf("%s\n", string(b))
	} else {
		err = ioutil.WriteFile(output, b, 0644)
		if err != nil {
			return errors.Wrapf(err, "Failed to write endpoints file to [%s]", output)
		}
	}

	log.WithFields(log.Fields{
		"num": len(endpoints),
		"output": output,
	}).Info("Written endpoints")
	return nil
}

func run(c *Config, ec2Client ec2iface.EC2API) {
	initLogging(c.Debug)
	log.WithField("version", version).Info("Portainer Endpoints")

	tag, err := NewTag(c.Tag)
	if err != nil {
		log.Fatal(err)
	}

	for {
		instances, err := getInstances(tag, ec2Client)
		if err != nil {
			log.Warnf("Error while fetching instances: %s", err)
			time.Sleep(c.Interval)
			continue
		}

		endpoints := []Endpoint{{
			Name: "local",
			URL:  "unix:///var/run/docker.sock",
		}}
		for _, i := range instances {
			endpoints = append(endpoints, i.GetEndpoint(c.Port))
		}

		err = writeEndpoints(endpoints, c.Output)
		if err != nil {
			log.Warnf("Error while writing endpoints: %s", err)
			time.Sleep(c.Interval)
			continue
		}

		time.Sleep(c.Interval)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "ddns"
	app.Usage = "Command line tool for dynamically generating domain entries for EC2 instances"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "tag, t",
			Usage:  "Tag used to filter EC2 instances. Format tag=value",
			EnvVar: envPrefix + "TAG",
		},
		cli.StringFlag{
			Name:   "output, o",
			Usage:  "Path of the output file",
			EnvVar: envPrefix + "OUTPUT",
		},
		cli.IntFlag{
			Name:   "port, p",
			Usage:  "Docker port",
			Value:  2375,
			EnvVar: envPrefix + "PORT",
		},
		cli.DurationFlag{
			Name:   "interval, i",
			Usage:  "Interval for querying the EC2 isntances",
			Value:  30 * time.Second,
			EnvVar: envPrefix + "INTERVAL",
		},
		cli.BoolFlag{
			Name:   "debug, D",
			Usage:  "Enable debug logging",
			EnvVar: envPrefix + "DEBUG",
		},
	}

	app.Action = func(c *cli.Context) error {
		run(&Config{
			Tag:      c.String("tag"),
			Output:   c.String("output"),
			Port:     c.Int("port"),
			Interval: c.Duration("interval"),
			Debug:    c.Bool("debug"),
		},
			NewEC2Client(),
		)
		return nil
	}

	app.Run(os.Args)
}
