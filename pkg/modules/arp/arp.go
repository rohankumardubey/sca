package arp

import (
	"net"
	"strings"

	"github.com/mdlayher/arp" //could be replace by https://github.com/google/gopacket/blob/master/examples/arpscan/arpscan.go
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
)

var argNoResolv bool

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
func (m *Module) New(options map[string]string) model.Module {
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
			if !argNoResolv {
				hosts, err := net.LookupAddr(arpPacket.SenderIP)
				if err == nil {
					v.Hostnames.Add(hosts...)
				}
			}
			table.Put(arp.SenderHardwareAddr.String(), v)
			out <- ModuleID
		}
	}
	return out
}

//Flags set for Module
func (m *Module) Flags() *pflag.FlagSet {
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

//GetData //TODO
func (m *Module) GetData() interface{} {
	return d.table.Values()
}
