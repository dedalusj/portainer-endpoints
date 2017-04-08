package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
	"time"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
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

	return string(body), nil
}

func getTagValue(tag string) (string, error) {
	instanceId, err := getInstanceId()
	if err != nil {
		return "", errors.Wrap(err, "Getting instance ID")
	}

	instance, err := getInstance(instanceId)
	if err != nil {
		return "", errors.Wrap(err, "Getting instance")
	}

	tagValue, ok := instance.Tags["Name"]
	if !ok {
		return "", fmt.Errorf("No tag [%s] present on instance [%s]", tag, instanceId)
	}

	return tagValue, nil
}

func prepareFilters(filters map[string][]*string) []*ec2.Filter {
	ec2Filter := []*ec2.Filter{}
	for k, v := range filters {
		ec2Filter = append(ec2Filter, &ec2.Filter{Name: &k, Values: v})
	}
	return ec2Filter
}

func getInstances(instanceIds []*string, filters map[string][]*string) ([]*Instance, error) {
        s := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String("ap-southeast-2")},
	}))
	svc := ec2.New(s)
	params := &ec2.DescribeInstancesInput{
		Filters: prepareFilters(filters),
		InstanceIds: instanceIds,
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return []*Instance{}, errors.Wrap(err, "Describing instances")
	}

	instances := []*Instance{}
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, NewInstance(i))
		}
	}
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
	tagKey := "tag:" + tag
	tagValue, err := getTagValue(tag)
	if err != nil {
		return []*Instance{}, errors.Wrap(err, "Getting tag value")
	}

	filters := map[string][]*string{
		tagKey: {&tagValue},
		"instance-state-name": {stringPtr("pending"), stringPtr("running")},
	}
	return getInstances([]*string{}, filters)
}
