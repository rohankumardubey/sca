package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sapk/sca/pkg"
	"github.com/sapk/sca/pkg/api"
	"github.com/sapk/sca/pkg/module"
)

//TODO use a config file
//TODO optimize data transfert to update only needed data
//TODO watch docker event

var (
	//Version version of running code
	version = "testing" // By default use testing but will be set at build time on release -X main.version=v${VERSION}
	commit  = "none"    // By default use none but will be set at build time on release -X main.commit=$(shell git log -q -1 | head -n 1 | cut -f2 -d' ')

	refreshToken string
	baseURL      string
	apiKey       string

	dockerEndpoint string

	timeout time.Duration

	cmd = &cobra.Command{
		Use:              "sca",
		Short:            "Simple Collector Agent",
		Long:             pkg.LongHelp,
		PersistentPreRun: pkg.SetupLogger,
	}
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display one-time collected informations in term for testing",
		Run: func(cmd *cobra.Command, args []string) {
			modules := module.GetList(getOptions())
			j, _ := json.MarshalIndent(getData(modules), "", "  ")
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
	cmd.PersistentFlags().StringVarP(&dockerEndpoint, pkg.EndpointFlag, "e", "unix:///var/run/docker.sock", "Docker endpoint.  Can also set default environment DOCKER_HOST")

	daemonCmd.Flags().DurationVarP(&timeout, pkg.TimeoutFlag, "r", 1*time.Minute, "Timeout before force refresh of collected data without event trigger during timeout period")
	daemonCmd.Flags().StringVarP(&refreshToken, pkg.TokenFlag, "t", "", "Firebase authentification token")
	daemonCmd.Flags().StringVarP(&baseURL, pkg.BaseURLFlag, "u", "", "Firebase base url")
	daemonCmd.Flags().StringVarP(&apiKey, pkg.APIFlag, "k", "", "Firebase api key")
	//TODO Setup a list modules to enable like modules=host,collector,docker ...
	//TODO add flag to force UUID
}

func startDaemon(cmd *cobra.Command, args []string) {
	api, err := api.New(apiKey, refreshToken, baseURL) //Init API
	if err != nil {
		log.Fatal("Fail to init API backend", err)
	}
	modules := module.GetList(getOptions())
	api.Send(getData(modules))

	c := time.Tick(timeout)
	for now := range c {
		log.Debug("Timeout tick triggered ", now)
		api.Send(getData(modules))
	}
	//func (c *Client) AddEventListener(listener chan<- *APIEvents) error
}

func getOptions() map[string]string {
	return map[string]string{
		"app.version":     version,
		"app.commit":      commit,
		"docker.endpoint": dockerEndpoint,
	}
}

func getData(modules map[string]module.Module) map[string]interface{} {
	d := make(map[string]interface{})
	for k, m := range modules {
		if m != nil {
			data := m.GetData()
			if structs.IsStruct(data) { //Object
				d[m.ID()] = structs.Map(data)
			} else { //String or something direct
				d[m.ID()] = data
			}
		} else {
			log.Debug("Skipping module ", k, " !")
		}
	}
	return d
}
