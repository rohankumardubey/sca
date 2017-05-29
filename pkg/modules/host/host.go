package host

import (
	"github.com/sapk-fork/spwd/proc"
	"github.com/sapk/sca/pkg/model"
	os "github.com/sapk/sca/pkg/modules/host/linux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

//ModuleID define the id of module
const ModuleID = "host"

//Module retrieve information form executing sca
type Module struct {
	Proc model.HostProc
}

//New constructor for Module
func New(options map[string]string) model.Module {
	log.WithFields(log.Fields{
		"id":      ModuleID,
		"options": options,
	}).Debug("Creating new Module")
	p := proc.ProcAll{}
	p.Init()
	return &Module{Proc: &p}
}

//Flags for Module
func Flags() *pflag.FlagSet {
	return nil
}

//ID return module ID
func (m *Module) ID() string {
	return ModuleID
}

//Event return event chan
func (m *Module) Event() <-chan string {
	return nil
}

//GetData //TODO
func (m *Module) GetData() interface{} {
	return os.GetData(m.Proc)
}
