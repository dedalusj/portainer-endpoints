package main

type Endpoint struct {
	Name      string
	URL       string
	TLS       bool
	TLSCACert string
	TLSCert   string
	TLSKey    string
}

func createEndpoints(instances []*Instance, config DockerConfig) []*Endpoint {
	endpoints := []*Endpoint{}
	for _, i := range instances {
		endpoint := &Endpoint{
			Name: i.Name(),
			URL: i.DockerURL(config.Port),
			TLS: config.TLS,
			TLSCACert: config.CACert,
			TLSCert: config.Cert,
			TLSKey: config.Key,
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

