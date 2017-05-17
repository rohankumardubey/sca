package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/eapache/channels"
	"github.com/sapk/sca/pkg"
	"github.com/sapk/sca/pkg/api"
	"github.com/sapk/sca/pkg/modules"
	"github.com/sapk/sca/pkg/tools"
)

//TODO use a config file
//TODO optimize data transfert to update only needed data
//TODO watch docker event

var (
	//Version version of running code
	version  = "testing" // By default use testing but will be set at build time on release -X main.version=v${VERSION}
	commit   = "none"    // By default use none but will be set at build time on release -X main.commit=$(shell git log -q -1 | head -n 1 | cut -f2 -d' ')
	dbFormat = "0"       //Used to evaluate compatibility with web UI

	refreshToken string
	baseURL      string
	apiKey       string
	moduleList   string

	timeout  time.Duration
	debounce = 1 * time.Second

	cmd = &cobra.Command{
		Use:              "sca",
		Short:            "Simple Collector Agent",
		PersistentPreRun: tools.SetupLogger,
	}
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display one-time collected informations in term for testing",
		Run: func(cmd *cobra.Command, args []string) {
			ms := modules.Create(getOptions())
			j, _ := json.MarshalIndent(ms.GetData(), "", "  ")
			fmt.Println(string(j))
		},
	}
	daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Start collecting informations and send them to the remote database",
		Run:   startDaemon,
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display current version and build date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s - Commit: %s\n", version, commit)
		},
	}
)

func main() {
	setupFlags()
	cmd.Long = fmt.Sprintf(pkg.LongHelp, version, commit)
	cmd.AddCommand(versionCmd, infoCmd, daemonCmd)
	cmd.Execute()
}

func setupFlags() {
	cmd.PersistentFlags().BoolP(pkg.VerboseFlag, "v", false, "Turns on verbose logging")
	cmd.PersistentFlags().StringVarP(&moduleList, pkg.ModulesFlag, "m", "", "Module list to load/enable. (--modules=host,collector,docker)")
	cmd.PersistentFlags().AddFlagSet(modules.Flags())

	daemonCmd.Flags().DurationVarP(&timeout, pkg.TimeoutFlag, "r", 5*time.Minute, "Timeout before force refresh of collected data without event trigger during timeout period")
	daemonCmd.Flags().StringVarP(&refreshToken, pkg.TokenFlag, "t", "", "Firebase authentification token")
	daemonCmd.Flags().StringVarP(&baseURL, pkg.BaseURLFlag, "u", "", "Firebase base url")
	daemonCmd.Flags().StringVarP(&apiKey, pkg.APIFlag, "k", "", "Firebase api key")
}

func startDaemon(cmd *cobra.Command, args []string) {
	api, err := api.New(apiKey, refreshToken, baseURL) //Init API
	if err != nil {
		log.Fatal("Fail to init API backend", err)
	}
	ms := modules.Create(getOptions())
	api.Send(ms.GetData())

	done := make(chan bool)
	go tools.Debounce(debounce, 20, tools.MergeChan(ms.Event(), channels.Wrap(time.Tick(timeout)).Out()), func(arg interface{}) {
		//Debounce in order to limit system call
		reason := "unkown"
		switch arg.(type) {
		case time.Time:
			reason = "timeout"
		case string:
			reason = arg.(string)
		}
		log.WithFields(log.Fields{
			"reason": reason, //Debug
			"arg":    arg,    //Debug
		}).Debug("Requesting fresh data to modules!")
		api.Send(ms.GetData()) //After a event or timeout get data and send it (with a debounce in case of recursive event)
	})
	<-done //Never end
}

func getOptions() map[string]string {
	return map[string]string{
		"app.version":  version,
		"app.commit":   commit,
		"app.dbFormat": dbFormat,
		"module.list":  moduleList,
	}
}
