package arp

import (
	"net"
	"strings"

	"github.com/mdlayher/arp"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
)

//Module retrieve information form executing sca
type Module struct {
	table *treemap.Map
	event <-chan string
}
type macEntry struct {
	IPs *treeset.Set
	Hostnames *treeset.Set
}

//ModuleID define the id of module
const ModuleID = "arp"

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")
	table := treemap.NewWithStringComparator()
	return &Module{table: table, event: setListener(table)}
}
func setListener(table *treemap.Map) <-chan string {
	c : arp.Dial(/*TODO*/)
	out := make(chan string)
	//TODO fill a table with all mac discovered
	go func() {
		for {	
			arpPacket, _, err := c.Read()
			if err != nil {
				log.Error(err)
				continue
			}

			if arpPacket.Operation != OperationReply {
				continue
			}
			v, ok := m.Get(arp.SenderHardwareAddr.String()) 
			if !ok {
				v = macEntry {
					IPs : treeset.NewWithStringComparator(),
					Hostnames : treeset.NewWithStringComparator(),
				}
			}
			
			v.IPs.Add(arpPacket.SenderIP)
			//TODO add module flags for resolv activation
			hosts, err := net.LookupAddr(arpPacket.SenderIP)
			if err == nil {
				v.Hostnames.Add(hosts...)
			}
			table.Put(arp.SenderHardwareAddr.String(), v)
			out <- ModuleID
		}
	}
	return out
}
//ID id of module
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return m.event
}

//GetData //TODO
func (m *Module) GetData() interface{} {
	d.table.Values()
	/* solution based on "github.com/mostlygeek/arp"
	t := arp.Table()
	r := make(Response, len(t))
	//hostList := make(map[string][]string, len(t))
	//macList := make(map[string]string, len(t))

	for ip, mac := range t {
		ipClean := strings.Replace(ip, ".", "-", -1)
		h := host{
			MAC: mac,
		}
		hosts, err := net.LookupAddr(ip)
		if err == nil {
			h.Hosts = hosts
		}
		r[ipClean] = h
	}
	return r
	*/
}
