package collector

import (
	"strings"
	"time"

	"github.com/sapk/sca/pkg/model"
	"github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
)

//ModuleID define the id of module
const ModuleID = "collector"

//Module retrieve information form executing sca
type Module struct {
	Version   string
	Commit    string
	DBFormat  string
	StartTime int64
	Config    map[string]string
	/* event     <-chan string */
}

//Response describe collector informations
type Response struct {
	Version    string            `json:"Version,omitempty"`
	Commit     string            `json:"Commit,omitempty"`
	DBFormat   string            `json:"DBFormat,omitempty"`
	StartTime  int64             `json:"StartTime,omitempty"`
	UpdateTime int64             `json:"UpdateTime,omitempty"`
	Status     collectorStatus   `json:"Status,omitempty"`
	Config     map[string]string `json:"Config,omitempty"`
}

//New constructor for Module
func (m *Module) New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{StartTime: time.Now().Unix(), Version: options["app.version"], DBFormat: options["app.dbFormat"], Commit: options["app.commit"], Config: getConfig(options) /* event: make(<-chan string)*/}
}

//Flagsset for Module
func (m *Module) Flags() *pflag.FlagSet {
	return nil
}

//ID return module ID
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return nil
	//return m.event
}

func getConfig(options map[string]string) map[string]string {
	tmp := make(map[string]string)
	for id, conf := range options {
		if strings.HasPrefix(id, "app.") {
			continue
		}
		tmp[strings.Replace(id, ".", "-", -1)] = conf
	}
	return tmp
}

//GetData //TODO
func (m *Module) GetData() interface{} {
	return Response{
		Version:    m.Version,
		Commit:     m.Commit,
		DBFormat:   m.DBFormat,
		StartTime:  m.StartTime,
		UpdateTime: time.Now().Unix(),
		Status:     getStatus(),
		Config:     m.Config,
	}
}
