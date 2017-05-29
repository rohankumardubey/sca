package arp

import (
	"net"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/mdlayher/arp" //could be replace by https://github.com/google/gopacket/blob/master/examples/arpscan/arpscan.go
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var argNoResolv bool

//Module retrieve information form executing sca
type Module struct {
	table *treemap.Map
	event chan string
}
type macEntry struct {
	IPs       *treeset.Set
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
	newm := &Module{table: treemap.NewWithStringComparator()}
	newm.setListener()
	return newm
}

func (m *Module) setListener() <-chan string {

	// Get a list of all interfaces.
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	out := make(chan string)
	for _, iface := range ifaces {
		c, err := arp.Dial(&iface)
		if err != nil {
			log.Warn(err)
		}
		go func(m *Module) {
			for {
				arpPacket, _, err := c.Read()
				if err != nil {
					log.Error(err)
					continue
				}

				if arpPacket.Operation != arp.OperationReply {
					continue
				}
				mac := arpPacket.SenderHardwareAddr.String()
				v, ok := m.table.Get(mac)
				if !ok {
					v = macEntry{
						IPs:       treeset.NewWithStringComparator(),
						Hostnames: treeset.NewWithStringComparator(),
					}
				}

				e := v.(macEntry)
				e.IPs.Add(arpPacket.SenderIP)
				if !argNoResolv {
					hosts, err := net.LookupAddr(arpPacket.SenderIP.String())
					if err == nil {
						for _, h := range hosts {
							e.Hostnames.Add(h)
						}
					}
				}
				log.WithFields(log.Fields{
					"mac":   mac,
					"value": e,
				}).Debug("Addding to arp mac list")
				m.table.Put(mac, e)
				m.event <- ModuleID
			}
		}(m)
	}
	return out
}

//Flags set for Module
func Flags() *pflag.FlagSet {
	fSet := pflag.NewFlagSet(ModuleID, pflag.ExitOnError)
	fSet.BoolVar(&argNoResolv, "arp-no-resolve", false, "resolve reverse-dns of ip found by arp. (default:false)")
	return fSet
}

//ID id of module
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return m.event
}

//GetData //TODO wait for atleast one result (maybe start a scan at startup ?)
func (m *Module) GetData() interface{} {
	return m.table.Values()
}
