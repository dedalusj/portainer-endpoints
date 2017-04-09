package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	log "github.com/Sirupsen/logrus"
	"time"
	"io/ioutil"
	"net/http"
)

func stringPtr(s string) *string { return &s }

type Instance struct {
	IpAddress string
	Tags      map[string]string
}

func NewInstance(ec2Instance *ec2.Instance) *Instance {
	ipAddress := *ec2Instance.PrivateIpAddress
	tags := make(map[string]string)
	for _, t := range ec2Instance.Tags {
		tags[*t.Key] = *t.Value
	}
	return &Instance{IpAddress: ipAddress, Tags: tags}
}

func (i *Instance) Name() string {
	name := "ip-" + strings.Replace(i.IpAddress, ".", "-", -1)
	if v, ok := i.Tags["Name"]; ok {
		name = v + "-" + name
	}
	return name
}

func (i *Instance) DockerURL(dockerPort int) string {
	return fmt.Sprintf("tcp://%s:%d/", i.IpAddress, dockerPort)
}

func getInstanceId() (string, error) {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		return "", errors.Wrap(err, "Querying EC2 metadata")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Request to metadata failed with code [%d]", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "Reading EC2 metadata query body")
	}

	instanceId := string(body)
	log.WithField("instanceId", instanceId).Debug("Found id for current instance")
	return string(body), nil
}

func getTagValue(tag string) (string, error) {
	instanceId, err := getInstanceId()
	if err != nil {
		return "", errors.Wrap(err, "Getting instance ID")
	}

	instance, err := getInstance(instanceId)
	if err != nil {
		return "", errors.Wrapf(err, "Getting info for instance id [%s]", instanceId)
	}

	tagValue, ok := instance.Tags[tag]
	if !ok {
		return "", fmt.Errorf("No tag [%s] present on instance [%s]", tag, instanceId)
	}

	log.WithFields(log.Fields{
		"name": tag,
		"value": tagValue,
	}).Debug("Found tag value for current instance")
	return tagValue, nil
}

func prepareFilters(filters map[string][]*string) []*ec2.Filter {
	ec2Filter := []*ec2.Filter{}
	for k, v := range filters {
		name := string(k)
		ec2Filter = append(ec2Filter, &ec2.Filter{Name: &name, Values: v})
	}
	return ec2Filter
}

func getInstances(instanceIds []*string, filters map[string][]*string) ([]*Instance, error) {
        s := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))},
	}))
	svc := ec2.New(s)

	params := &ec2.DescribeInstancesInput{}
	if len(instanceIds) > 0 {
		params.InstanceIds = instanceIds
	} else if len(filters) > 0 {
		params.Filters = prepareFilters(filters)
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return []*Instance{}, errors.Wrapf(err, "Describing instances params=[%s]", params)
	}

	instances := []*Instance{}
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, NewInstance(i))
		}
	}

	log.WithFields(log.Fields{
		"number": len(instances),
		"instanceIDs": awsutil.Prettify(instanceIds),
		"filters": awsutil.Prettify(filters),
	}).Debug("Fetched instances")

	return instances, nil
}

func getInstance(instanceId string) (*Instance, error) {
	instances, err := getInstances([]*string{&instanceId}, map[string][]*string{})
	if err != nil {
		return &Instance{}, errors.Wrap(err, "Getting instances")
	}
	return instances[0], nil
}

func getFilteredInstances(tag string) ([]*Instance, error) {
	log.WithField("tag", tag).Debug("Fetching instances with tag")

	tagKey := "tag:" + tag
	tagValue, err := getTagValue(tag)
	if err != nil {
		return []*Instance{}, errors.Wrapf(err, "Getting value for tag [%s]", tag)
	}

	filters := map[string][]*string{
		tagKey: {&tagValue},
		"instance-state-name": {stringPtr("pending"), stringPtr("running")},
	}
	return getInstances([]*string{}, filters)
}
