package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

const (
	//VerboseFlag flag to set more verbose level
	VerboseFlag = "verbose"
	//EndpointFlag flag to set the endpoint to use (default: unix:///var/run/docker.sock)
	EndpointFlag = "endpoint"
	//EndpointEnv env to set endpoint of docker
	EndpointEnv = "DOCKER_HOST"

	longHelp = `
sca (Simple Collector Agent)
Collect local data and forward then to a realtime database.
`
)

var (
	client *docker.Client
	cmd    = &cobra.Command{
		Use:              "sca",
		Short:            "Simple Collector Agent",
		Long:             longHelp,
		PersistentPreRun: setupLogger,
		Run:              start,
	}
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display one-time collected informations in term for testing",
		Run: func(cmd *cobra.Command, args []string) {
			client := initClient(cmd, args)
			j, _ := json.MarshalIndent(getData(client), "", "  ")
			fmt.Println(string(j))
		},
	}
)

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
	Networks   []docker.Network       `json:"Networks"`
}

//GlobalResponse object json
type GlobalResponse struct {
	Host   *HostResponse   `json:"Host"`
	Docker *DockerResponse `json:"Docker"`
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
}

func main() {
	setupFlags()
	cmd.AddCommand(infoCmd)
	cmd.Execute()
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
	imgs, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		panic(err)
	}
	nets, err := client.ListNetworks()
	if err != nil {
		panic(err)
	}
	cnts, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		panic(err)
	}
	info, err := client.Info()
	if err != nil {
		panic(err)
	}
	return &DockerResponse{
		Info:       info,
		Containers: cnts,
		Images:     imgs,
		Networks:   nets,
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
	//var ints [len(ifaces)]InterfaceResponse
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
		Host:   getHostData(client),
		Docker: getDockerData(client),
	}
}
func start(cmd *cobra.Command, args []string) {
	client := initClient(cmd, args)
	j, _ := json.Marshal(getData(client))
	log.Debugln(string(j))
	//TODO monitor event and update data
	//func (c *Client) AddEventListener(listener chan<- *APIEvents) error
}
