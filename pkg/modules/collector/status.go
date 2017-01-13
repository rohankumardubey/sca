package collector

import (
	"fmt"
	"runtime"
	"time"
)

type collectorStatus struct {
	NumGoroutine int

	// General statistics.
	MemAllocated uint64 // bytes allocated and still in use
	MemTotal     uint64 // bytes allocated (even if freed)
	MemSys       uint64 // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 // number of pointer lookups
	MemMallocs   uint64 // number of mallocs
	MemFrees     uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    uint64 // bytes allocated and still in use
	HeapSys      uint64 // bytes obtained from system
	HeapIdle     uint64 // bytes in idle spans
	HeapInuse    uint64 // bytes in non-idle span
	HeapReleased uint64 // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  uint64 // bootstrap stacks
	StackSys    uint64
	MSpanInuse  uint64 // mspan structures
	MSpanSys    uint64
	MCacheInuse uint64 // mcache structures
	MCacheSys   uint64
	BuckHashSys uint64 // profiling bucket hash table
	GCSys       uint64 // GC metadata
	OtherSys    uint64 // other system allocations

	// Garbage collector statistics.
	NextGC       uint64 // next run in HeapAlloc time (bytes)
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

	sys.MemAllocated = m.Alloc
	sys.MemTotal = m.TotalAlloc
	sys.MemSys = m.Sys
	sys.Lookups = m.Lookups
	sys.MemMallocs = m.Mallocs
	sys.MemFrees = m.Frees

	sys.HeapAlloc = m.HeapAlloc
	sys.HeapSys = m.HeapSys
	sys.HeapIdle = m.HeapIdle
	sys.HeapInuse = m.HeapInuse
	sys.HeapReleased = m.HeapReleased
	sys.HeapObjects = m.HeapObjects

	sys.StackInuse = m.StackInuse
	sys.StackSys = m.StackSys
	sys.MSpanInuse = m.MSpanInuse
	sys.MSpanSys = m.MSpanSys
	sys.MCacheInuse = m.MCacheInuse
	sys.MCacheSys = m.MCacheSys
	sys.BuckHashSys = m.BuckHashSys
	sys.GCSys = m.GCSys
	sys.OtherSys = m.OtherSys

	sys.NextGC = m.NextGC
	sys.LastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	sys.PauseTotalNs = fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000)
	sys.PauseNs = fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	sys.NumGC = m.NumGC

	return sys
}
