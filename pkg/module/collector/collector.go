package collector

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sapk/sca/pkg/tool"

	log "github.com/sirupsen/logrus"
)

const id = "Collector"

//Module retrieve information form executing sca
type Module struct {
	Version   string
	Commit    string
	DBFormat  string
	StartTime int64
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
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")
	return &Module{StartTime: time.Now().Unix(), Version: options["app.version"], DBFormat: options["app.dbFormat"], Commit: options["app.commit"]}
}

//ID //TODO
func (c *Module) ID() string {
	return id
}

//GetData //TODO
func (c *Module) GetData() interface{} {
	return Response{
		Version:    c.Version,
		Commit:     c.Commit,
		DBFormat:   c.DBFormat,
		StartTime:  c.StartTime,
		UpdateTime: time.Now().Unix(),
		Status:     getStatus(),
	}
}

type collectorStatus struct {
	NumGoroutine int

	// General statistics.
	MemAllocated string // bytes allocated and still in use
	MemTotal     string // bytes allocated (even if freed)
	MemSys       string // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 // number of pointer lookups
	MemMallocs   uint64 // number of mallocs
	MemFrees     uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string // bytes allocated and still in use
	HeapSys      string // bytes obtained from system
	HeapIdle     string // bytes in idle spans
	HeapInuse    string // bytes in non-idle span
	HeapReleased string // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  string // bootstrap stacks
	StackSys    string
	MSpanInuse  string // mspan structures
	MSpanSys    string
	MCacheInuse string // mcache structures
	MCacheSys   string
	BuckHashSys string // profiling bucket hash table
	GCSys       string // GC metadata
	OtherSys    string // other system allocations

	// Garbage collector statistics.
	NextGC       string // next run in HeapAlloc time (bytes)
	LastGC       string // last run in absolute time (ns)
	PauseTotalNs string
	PauseNs      string // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32
}

func getStatus() collectorStatus {
	var sys collectorStatus

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	sys.NumGoroutine = runtime.NumGoroutine()

	sys.MemAllocated = tool.FileSize(int64(m.Alloc))
	sys.MemTotal = tool.FileSize(int64(m.TotalAlloc))
	sys.MemSys = tool.FileSize(int64(m.Sys))
	sys.Lookups = m.Lookups
	sys.MemMallocs = m.Mallocs
	sys.MemFrees = m.Frees

	sys.HeapAlloc = tool.FileSize(int64(m.HeapAlloc))
	sys.HeapSys = tool.FileSize(int64(m.HeapSys))
	sys.HeapIdle = tool.FileSize(int64(m.HeapIdle))
	sys.HeapInuse = tool.FileSize(int64(m.HeapInuse))
	sys.HeapReleased = tool.FileSize(int64(m.HeapReleased))
	sys.HeapObjects = m.HeapObjects

	sys.StackInuse = tool.FileSize(int64(m.StackInuse))
	sys.StackSys = tool.FileSize(int64(m.StackSys))
	sys.MSpanInuse = tool.FileSize(int64(m.MSpanInuse))
	sys.MSpanSys = tool.FileSize(int64(m.MSpanSys))
	sys.MCacheInuse = tool.FileSize(int64(m.MCacheInuse))
	sys.MCacheSys = tool.FileSize(int64(m.MCacheSys))
	sys.BuckHashSys = tool.FileSize(int64(m.BuckHashSys))
	sys.GCSys = tool.FileSize(int64(m.GCSys))
	sys.OtherSys = tool.FileSize(int64(m.OtherSys))

	sys.NextGC = tool.FileSize(int64(m.NextGC))
	sys.LastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	sys.PauseTotalNs = fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000)
	sys.PauseNs = fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	sys.NumGC = m.NumGC

	return sys
}
