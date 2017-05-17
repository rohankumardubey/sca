package uuid

import (
	"os"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
)

const ModuleID = "UUID"

//Module retrieve information form executing sca
type Module struct {
	UUID string
}

//Response describe collector informations
type Response string

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")
	hostname, err := os.Hostname()
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
	return &Module{UUID: u5.String()} //TODO use option to get a user or config (/etc/sca/uuid or via cmd ?) defined uuid
}
//ID //TODO
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return nil
}

//GetData //TODO
func (m *Module) GetData() interface{} {
	return m.UUID
}
