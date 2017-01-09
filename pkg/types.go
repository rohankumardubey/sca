package pkg

import docker "github.com/fsouza/go-dockerclient"

//DockerResponse describe a docker host informations
type DockerResponse struct {
	Info       *docker.DockerInfo     `json:"Info,omitempty"`
	Containers []docker.APIContainers `json:"Containers,omitempty"`
	Images     []docker.APIImages     `json:"Images,omitempty"`
	Volumes    []docker.Volume        `json:"Volumes,omitempty"`
	Networks   []docker.Network       `json:"Networks,omitempty"`
}

/*
//GlobalResponse object json
type GlobalResponse struct {
	UUID      string             `json:"UUID,omitempty"`
	Host      *HostResponse      `json:"Host,omitempty"`
	Collector *CollectorResponse `json:"Collector,omitempty"`
	Docker    *DockerResponse    `json:"Docker,omitempty"` //TODO add information on collector version
}
*/
