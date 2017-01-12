package linux

import (
	"net"
	"os"

	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
)

//GetData Return host information of linux os
func GetData(p model.HostProc) interface{} {
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
	ints := make([]model.HostInterfaceResponse, len(ifaces))
	for id, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.WithFields(log.Fields{
				"iface": i,
				"addrs": addrs,
				"err":   err,
			}).Warn("Failed to retrieve addrs of interfaces")
		}
		ints[id] = model.HostInterfaceResponse{
			Info:  i,
			Addrs: addrs,
		}
	}
	p.Update()
	return model.HostResponse{
		Name:       hostname,
		Interfaces: ints,
		Proc:       p,
	}
}
