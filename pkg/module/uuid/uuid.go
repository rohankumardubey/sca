package uuid

import (
	"os"

	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
)

const id = "UUID"

//Module retrieve information form executing sca
type Module struct {
	UUID string
}

//Response describe collector informations
type Response string

//New constructor for Module
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	hostname, err := os.Hostname() //TODO maybe cache it at build time ?
	if err != nil {
		log.WithFields(log.Fields{
			"hostname": hostname,
			"err":      err,
		}).Warn("Failed to retrieve hostname")
	}
	u5, err := uuid.NewV5(uuid.NamespaceURL, []byte(hostname)) //TODO better discriminate maybe add time and save it in /etc/sca/uuid ?
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": u5,
			"err":  err,
		}).Fatal("Failed to generate uuid")
	}
	return &Module{UUID: u5.String()} //TODO use option to get a user or config (/etc/sca/uuid ?) defined uuid
}

//ID //TODO
func (m *Module) ID() string {
	return id
}

//GetData //TODO
func (m *Module) GetData() interface{} {
	return m.UUID
}
