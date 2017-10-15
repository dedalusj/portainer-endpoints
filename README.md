DDNS
====

[![Build Status](https://travis-ci.org/dedalusj/portainer-endpoints.svg?branch=master)](https://travis-ci.org/dedalusj/portainer-endpoints)

Portainer endpoints is a command line tool that dynamically maintains a portainer endpoints file with running EC2 instances.

It does so by querying all pending and running EC2 instances with a specified tag and write them to a specified json file.

#### Command Line Arguments

- `--tag`: Specify the tag and value to use when querying for EC2 instances. Format `tag=value`.
- `--output`: Output path where the portainer endpoints file will be written.
- `--port`: Docker remote API port. Default `2375`.
- `--interval`: Interval for querying for EC2 instances. See [https://golang.org/pkg/time/#ParseDuration](https://golang.org/pkg/time/#ParseDuration) for format. Default `30s`.
- `--debug`: Enable debug logging.

The command line parameters can also be controlled with an environment variable of the form `PE_<parameter_name>`.

#### Development

Portainer endpoints relies on [dep](https://github.com/golang/dep) to version its dependencies.

To build portainer endpoints run `make build`. To build the docker image run `make docker`