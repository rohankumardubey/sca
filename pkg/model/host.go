package model

import "net"

//HostProc //proc.ProcAll
type HostProc interface {
	Init()
	Update()
}

//HostResponse describe hots informations
type HostResponse struct {
	Name       string                           `json:"Name,omitempty"`
	Interfaces map[string]HostInterfaceResponse `json:"Interfaces,omitempty"`
	Proc       HostProc                         `json:"Proc,omitempty"`
	//TODO add ressources IP, CPU, MEM
}

//HostInterfaceResponse describe interface
type HostInterfaceResponse struct {
	//HWAddr string        `json:"HWAddr,omitempty"`
	Info  net.Interface `json:"Info,omitempty"` //TODO add ressources IP, CPU, MEM
	Addrs []net.Addr    `json:"Addrs,omitempty"`
}
