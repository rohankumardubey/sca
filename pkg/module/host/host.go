package host

import (
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

const id = "Host"

//Module retrieve information form executing sca
type Module struct {
}

//Response describe hots informations
type Response struct {
	Name       string              `json:"Name,omitempty"`
	Interfaces []InterfaceResponse `json:"Interfaces,omitempty"`
	//TODO add ressources IP, CPU, MEM
}

//InterfaceResponse describe interface informations
type InterfaceResponse struct {
	Info  net.Interface `json:"Info,omitempty"` //TODO add ressources IP, CPU, MEM
	Addrs []net.Addr    `json:"Addrs,omitempty"`
}

//New constructor for Module
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{}
}

//ID //TODO
func (c *Module) ID() string {
	return id
}

//GetData //TODO
func (c *Module) GetData() interface{} {
	hostname, err := os.Hostname() //TODO maybe cache it at build time ?
	if err != nil {
		log.WithFields(log.Fields{
			"hostname": hostname,
			"err":      err,
		}).Warn("Failed to retrieve hostname")
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		log.WithFields(log.Fields{
			"ifaces": ifaces,
			"err":    err,
		}).Warn("Failed to retrieve host interfaces")
	}
	ints := make([]InterfaceResponse, len(ifaces))
	for id, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.WithFields(log.Fields{
				"iface": i,
				"addrs": addrs,
				"err":   err,
			}).Warn("Failed to retrieve addrs of interfaces")
		}
		ints[id] = InterfaceResponse{
			Info:  i,
			Addrs: addrs,
		}
	}
	return Response{
		Name:       hostname,
		Interfaces: ints,
	}
}
