package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

const (
	//Version version of running code
	Version = "testing" // By default use testing but will be set at build time on release -X main.version=v${VERSION}
	//VerboseFlag flag to set more verbose level
	VerboseFlag = "verbose"
	//EndpointFlag flag to set the endpoint to use (default: unix:///var/run/docker.sock)
	EndpointFlag = "endpoint"
	//EndpointEnv env to set endpoint of docker
	EndpointEnv = "DOCKER_HOST"
	//TimeoutFlag flag to set timeout period
	TimeoutFlag = "timeout"
	longHelp    = `
sca (Simple Collector Agent)
Collect local data and forward them to a realtime database.
== Version: %s ==
`
)

var (
	timeout   time.Duration
	startTime = time.Now()
	client    *docker.Client
	cmd       = &cobra.Command{
		Use:              "sca",
		Short:            "Simple Collector Agent",
		Long:             longHelp,
		PersistentPreRun: setupLogger,
	}
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display one-time collected informations in term for testing",
		Run: func(cmd *cobra.Command, args []string) {
			client := initClient(cmd, args)
			j, _ := json.MarshalIndent(getData(client), "", "  ")
			//j, _ := json.Marshal(getData(client))
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
			fmt.Printf("Version: %s\n", Version)
		},
	}
)

//CollectorResponse describe collector informations
type CollectorResponse struct {
	Version   string    `json:"Version"`
	StartTime time.Time `json:"StartTime"` //TODO use a date object
}

//HostResponse describe host informations
type HostResponse struct {
	Name       string              `json:"Name"` //TODO add ressources IP, CPU, MEM
	Interfaces []InterfaceResponse `json:"Interfaces"`
}

//InterfaceResponse describe interface informations
type InterfaceResponse struct {
	Info  net.Interface `json:"Info"` //TODO add ressources IP, CPU, MEM
	Addrs []net.Addr    `json:"Addrs"`
}

//DockerResponse describe a docker host informations
type DockerResponse struct {
	Info       *docker.DockerInfo     `json:"Info"`
	Containers []docker.APIContainers `json:"Containers"`
	Images     []docker.APIImages     `json:"Images"`
	Volumes    []docker.Volume        `json:"Volumes"`
	Networks   []docker.Network       `json:"Networks"`
}

//GlobalResponse object json
type GlobalResponse struct {
	Host      *HostResponse      `json:"Host"`
	Collector *CollectorResponse `json:"Collector"`
	Docker    *DockerResponse    `json:"Docker"` //TODO add information on collector version
}

func main() {
	setupFlags()
	cmd.Long = fmt.Sprintf(longHelp, Version)
	cmd.AddCommand(versionCmd, infoCmd, daemonCmd)
	cmd.Execute()
}

func typeOrEnv(cmd *cobra.Command, flag, envname string) string {
	val, _ := cmd.Flags().GetString(flag)
	if val == "" {
		val = os.Getenv(envname)
	}
	return val
}

func setupLogger(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool(VerboseFlag); verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func setupFlags() {
	cmd.PersistentFlags().BoolP(VerboseFlag, "v", false, "Turns on verbose logging")
	cmd.PersistentFlags().StringP(EndpointFlag, "e", "unix:///var/run/docker.sock", "Docker endpoint.  Can also set default environment DOCKER_HOST")

	daemonCmd.Flags().DurationVarP(&timeout, TimeoutFlag, "t", 5*time.Minute, "Timeout before force refresh of collected data without event trigger during timeout period")
	//TODO Setup a list modules to load like modules=host,collector,docker ...
}

func initClient(cmd *cobra.Command, args []string) *docker.Client {
	//TODO detect if remote and SSL
	endpoint := typeOrEnv(cmd, EndpointFlag, EndpointEnv)
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	return client
}
func getDockerData(client *docker.Client) *DockerResponse {
	//Get images
	imgs, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		panic(err)
	}
	//Get networks
	nets, err := client.ListNetworks()
	if err != nil {
		panic(err)
	}
	//Get container
	cnts, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		panic(err)
	}
	//Get volumes
	vols, err := client.ListVolumes(docker.ListVolumesOptions{})
	if err != nil {
		panic(err)
	}
	//Get server info
	info, err := client.Info()
	if err != nil {
		panic(err)
	}
	return &DockerResponse{
		Info:       info,
		Containers: cnts,
		Images:     imgs,
		Volumes:    vols,
		Networks:   nets,
	}
}
func getCollectorData() *CollectorResponse {
	return &CollectorResponse{
		Version:   Version,
		StartTime: startTime,
	}
}
func getHostData(client *docker.Client) *HostResponse {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	ints := make([]InterfaceResponse, len(ifaces), len(ifaces))
	for id, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		ints[id] = InterfaceResponse{
			Info:  i,
			Addrs: addrs,
		}
	}
	return &HostResponse{
		Name:       hostname,
		Interfaces: ints,
	}
}
func getData(client *docker.Client) *GlobalResponse {
	//TODO detect if docket
	return &GlobalResponse{
		Host:      getHostData(client),
		Collector: getCollectorData(),
		Docker:    getDockerData(client),
	}
}
func startDaemon(cmd *cobra.Command, args []string) {
	//TODO generate UUID to get persistance run ?
	//TODO monitor event and update data
	client := initClient(cmd, args)
	j, _ := json.Marshal(getData(client))
	log.Debugln(string(j))
	c := time.Tick(timeout)
	for now := range c {
		j, _ := json.Marshal(getData(client))
		//fmt.Printf("%v %s\n", now, string(j))
		log.Debugln(now, string(j))
	}
	//func (c *Client) AddEventListener(listener chan<- *APIEvents) error
}
