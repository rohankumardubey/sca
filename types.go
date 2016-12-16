package main

import (
	"net"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

//CollectorResponse describe collector informations
type CollectorResponse struct {
	Version   string    `json:"Version,omitempty"`
	StartTime time.Time `json:"StartTime,omitempty"`
	Hash      string    `json:"Hash,omitempty"`
}

//HostResponse describe host informations
type HostResponse struct {
	Name       string              `json:"Name,omitempty"` //TODO add ressources IP, CPU, MEM
	Interfaces []InterfaceResponse `json:"Interfaces,omitempty"`
}

//InterfaceResponse describe interface informations
type InterfaceResponse struct {
	Info  net.Interface `json:"Info,omitempty"` //TODO add ressources IP, CPU, MEM
	Addrs []net.Addr    `json:"Addrs,omitempty"`
}

//DockerResponse describe a docker host informations
type DockerResponse struct {
	Info       *docker.DockerInfo      `json:"Info,omitempty"`
	Containers *[]docker.APIContainers `json:"Containers,omitempty"`
	Images     *[]docker.APIImages     `json:"Images,omitempty"`
	Volumes    *[]docker.Volume        `json:"Volumes,omitempty"`
	Networks   *[]docker.Network       `json:"Networks,omitempty"`
}

//GlobalResponse object json
type GlobalResponse struct {
	UUID      string             `json:"UUID,omitempty"`
	Host      *HostResponse      `json:"Host,omitempty"`
	Collector *CollectorResponse `json:"Collector,omitempty"`
	Docker    *DockerResponse    `json:"Docker,omitempty"` //TODO add information on collector version
}
