package arp

import (
	"net"
	"strings"

	"github.com/mostlygeek/arp"
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
)

//Module retrieve information form executing sca
type Module struct {
}

//Response describe returned informations
type Response map[string]host

type host struct {
	MAC   string
	Hosts []string
}

//ModuleID define the id of module
const ModuleID = "arp"

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")

	//arp.AutoRefresh(5 * time.Minute)
	arp.CacheUpdate()
	return &Module{}
}

//ID id of module
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return nil
}

//GetData //TODO
func (m *Module) GetData() interface{} {
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
}
