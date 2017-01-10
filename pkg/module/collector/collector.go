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
	StartTime int64
}

//Response describe collector informations
type Response struct {
	Version    string `json:"Version,omitempty"`
	StartTime  int64  `json:"StartTime,omitempty"`
	UpdateTime int64  `json:"UpdateTime,omitempty"`
	Commit     string `json:"Commit,omitempty"`
}

//New constructor for Module
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{StartTime: time.Now().Unix(), Version: options["app.version"], Commit: options["app.commit"]}
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
		UpdateTime: time.Now().Unix(),
		Commit:     c.Commit,
	}
}
