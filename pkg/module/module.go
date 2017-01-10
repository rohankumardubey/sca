package module

import (
	"github.com/sapk/sca/pkg/module/collector"
	"github.com/sapk/sca/pkg/module/docker"
	"github.com/sapk/sca/pkg/module/host"
	"github.com/sapk/sca/pkg/module/uuid"
)

//Module represente a module collecting data
type Module interface {
	ID() string
	GetData() interface{}
}

//GetList Return module list initalized
func GetList(options map[string]string) map[string]Module {
	m := make(map[string]Module)
	//ms := make([]Module, 3, 100) //TODO dynamic loading
	m["Collector"] = collector.New(options)
	m["Host"] = host.New(options)
	m["Docker"] = docker.New(options)
	m["UUID"] = uuid.New(options)
	return m
}
