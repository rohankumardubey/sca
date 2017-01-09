package module

import (
	"github.com/sapk/sca/pkg/module/collector"
	"github.com/sapk/sca/pkg/module/host"
)

//Module represente a module collecting data
type Module interface {
	ID() string
	GetData() interface{}
}

//GetList Return module list initalized
func GetList(options map[string]string) []Module {
	ms := make([]Module, 2, 100) //TODO dynamic loading
	ms[0] = collector.New(options)
	ms[1] = host.New(options)
	return ms
}
