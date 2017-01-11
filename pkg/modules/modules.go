package modules

import (
	"github.com/eapache/channels"
	"github.com/fatih/structs"
	"github.com/sapk/sca/pkg/model"
	"github.com/sapk/sca/pkg/modules/collector"
	"github.com/sapk/sca/pkg/modules/docker"
	"github.com/sapk/sca/pkg/modules/host"
	"github.com/sapk/sca/pkg/modules/uuid"
	"github.com/sapk/sca/pkg/tools"
	log "github.com/sirupsen/logrus"
)

var (
	listModulesConstructor = []func(map[string]string) model.Module{collector.New, docker.New, host.New, uuid.New}
)

//ModuleList represente a module list
type ModuleList struct {
	list  map[string]model.Module
	event <-chan interface{}
}

//Create a module list and init them
func Create(options map[string]string) *ModuleList {
	list := getList(options)
	c := make([]<-chan interface{}, len(list))
	i := 0
	for _, m := range list {
		c[i] = channels.Wrap(m.Event()).Out()
		i++
	}
	return &ModuleList{
		list:  list,
		event: tools.MergeChan(c...),
	}
}

//getList Return module list initalized
func getList(options map[string]string) map[string]model.Module {
	m := make(map[string]model.Module, len(listModulesConstructor)) //TODO only load base on options args
	for _, fc := range listModulesConstructor {
		module := fc(options)
		m[module.ID()] = module
	}
	return m
}

//GetData request every module for a
func (ml *ModuleList) GetData() map[string]interface{} {
	d := make(map[string]interface{})
	for k, m := range ml.list {
		if m != nil {
			data := m.GetData()
			if structs.IsStruct(data) { //Object
				d[m.ID()] = structs.Map(data)
			} else { //String or something direct
				d[m.ID()] = data
			}
		} else {
			log.Debug("Skipping empty module ", k, " !")
		}
	}
	return d
}

//Event return event chan
func (ml *ModuleList) Event() <-chan interface{} {
	return ml.event
}
