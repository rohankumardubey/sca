package pkg

const (
	//VerboseFlag flag to set more verbose level
	VerboseFlag = "verbose"
	//TimeoutFlag flag to set timeout period
	TimeoutFlag = "timeout"
	//TokenFlag flag to set firebase token
	TokenFlag = "token"
	//APIFlag flag to set firebase api key
	APIFlag = "api"
	//BaseURLFlag flag to set firebase url
	BaseURLFlag = "url"
	//ModulesFlag flag to set module list to load/enable.
	ModulesFlag = "modules"
	//LongHelp help message of cmd
	LongHelp = `
sca (Simple Collector Agent)
Collect local data and forward them to a realtime database.
== Version: %s - Hash: %s ==
`
)
