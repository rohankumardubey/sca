package uuid

import (
	"os"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/sapk/sca/pkg/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

//ModuleID define the id of module
const ModuleID = "uuid"

var argUUID string

//Module retrieve information form executing sca
type Module struct {
	UUID string
}

//Response describe collector informations
type Response string

//Flags set for Module
func Flags() *pflag.FlagSet {
	fSet := pflag.NewFlagSet(ModuleID, pflag.ExitOnError)
	fSet.StringVar(&argUUID, "uuid", "", "uuid to use by this collector")
	return fSet
}

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")
	if argUUID == "" {
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
		argUUID = u5.String()
	}
	return &Module{UUID: argUUID}
}

//ID id of module
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return nil
}

//GetData data of module
func (m *Module) GetData() interface{} {
	return m.UUID
}
