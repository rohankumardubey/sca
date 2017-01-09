package collector

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const id = "Collector"

//Module retrieve information form executing sca
type Module struct {
	Version   string
	Commit    string
	StartTime time.Time
}

//Response describe collector informations
type Response struct {
	Version    string    `json:"Version,omitempty"`
	StartTime  time.Time `json:"StartTime,omitempty"`
	UpdateTime time.Time `json:"UpdateTime,omitempty"`
	Commit     string    `json:"Commit,omitempty"`
}

//New constructor for CollectorModule
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{StartTime: time.Now(), Version: options["version"], Commit: options["commit"]}
}

//ID //TODO
func (c *Module) ID() string {
	return id
}

//GetData //TODO
func (c *Module) GetData() interface{} {
	return Response{
		Version:    c.Version,
		StartTime:  c.StartTime,
		UpdateTime: time.Now(),
		Commit:     c.Commit,
	}
}
