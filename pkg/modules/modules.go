package modules

import (
	"strings"

	"github.com/sapk/sca/pkg/model"
	"github.com/sapk/sca/pkg/modules/arp"
	"github.com/sapk/sca/pkg/modules/collector"
	"github.com/sapk/sca/pkg/modules/docker"
	"github.com/sapk/sca/pkg/modules/host"
	"github.com/sapk/sca/pkg/modules/uuid"
	"github.com/sapk/sca/pkg/tools"

	"github.com/eapache/channels"
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	listModulesConstructor = map[string]func(map[string]string) model.Module{
		collector.ModuleID: collector.New,
		docker.ModuleID:    docker.New,
		host.ModuleID:      host.New,
		uuid.ModuleID:      uuid.New,
		arp.ModuleID:       arp.New,
	} //TODO use golang module format and separate code.
)

//ModuleList represente a module list
type ModuleList struct {
	list  map[string]model.Module
	event <-chan interface{}
}

//Flags set for Module
func Flags() *pflag.FlagSet {
	fSet := pflag.NewFlagSet("", pflag.ExitOnError)
	fSet.AddFlagSet(uuid.Flags())
	fSet.AddFlagSet(docker.Flags())
	//TODO add others modules and loop.
	return fSet
}

//Create a module list and init them
func Create(options map[string]string) *ModuleList {
	list := getList(options)
	c := make([]<-chan interface{}, len(list))
	i := 0
	for _, m := range list {
		ch := m.Event()
		if ch != nil {
			c[i] = channels.Wrap(m.Event()).Out()
			i++
		}
	}
	return &ModuleList{
		list:  list,
		event: tools.MergeChan(c...), //TODO test replace by github.com/eapache/channels.Multiplex
	}
}

//parseModuleListOption Return module list based on --modules list arg
func parseModuleListOption(options map[string]string) map[string]func(map[string]string) model.Module {
	if options["module.list"] == "" {
		return listModulesConstructor
	}

	mList := strings.Split(options["module.list"]+",uuid", ",") //Add uuid by force
	mContructors := make(map[string]func(map[string]string) model.Module, len(mList))
	for _, mName := range mList {
		mc, ok := listModulesConstructor[mName]
		if !ok {
			log.Fatalf("Module %s not found", mName) //TODO be more gracefull ^^
		}
		mContructors[mName] = mc //This is also removing duplicate.
	}
	return mContructors
}

//getList Return module list initalized
func getList(options map[string]string) map[string]model.Module {
	constructors := parseModuleListOption(options)
	modules := make(map[string]model.Module, len(constructors))
	for _, fInit := range constructors {
		module := fInit(options) //TODO only pass module.(module.ID()).xxx options
		modules[module.ID()] = module
	}
	return modules
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
