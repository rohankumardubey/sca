package collector

import (
	"time"

	"github.com/sapk/sca/pkg/model"

	log "github.com/sirupsen/logrus"
)

const id = "Collector"

//Module retrieve information form executing sca
type Module struct {
	Version   string
	Commit    string
	DBFormat  string
	StartTime int64
	event     <-chan string
}

//Response describe collector informations
type Response struct {
	Version    string          `json:"Version,omitempty"`
	Commit     string          `json:"Commit,omitempty"`
	DBFormat   string          `json:"DBFormat,omitempty"`
	StartTime  int64           `json:"StartTime,omitempty"`
	UpdateTime int64           `json:"UpdateTime,omitempty"`
	Status     collectorStatus `json:"Status,omitempty"`
}

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{StartTime: time.Now().Unix(), Version: options["app.version"], DBFormat: options["app.dbFormat"], Commit: options["app.commit"], event: make(<-chan string)}
}

//ID //TODO
func (m *Module) ID() string {
	return id
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return m.event
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
	}
}
