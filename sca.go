package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/structs"
	docker "github.com/fsouza/go-dockerclient"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//TODO use a config file
//TODO optimize data transfert to update only needed data
//TODO watch docker event

const (
	//VerboseFlag flag to set more verbose level
	VerboseFlag = "verbose"
	//EndpointFlag flag to set the endpoint to use (default: unix:///var/run/docker.sock)
	EndpointFlag = "endpoint"
	//EndpointEnv env to set endpoint of docker
	EndpointEnv = "DOCKER_HOST"
	//TimeoutFlag flag to set timeout period
	TimeoutFlag = "timeout"
	//TokenFlag flag to set firebase token
	TokenFlag = "token"
	//BaseURLFlag flag to set firebase url
	BaseURLFlag = "url"
	longHelp    = `
sca (Simple Collector Agent)
Collect local data and forward them to a realtime database.
== Version: %s - Hash: %s ==
`
)

var (
	//Version version of running code
	version = "testing" // By default use testing but will be set at build time on release -X main.version=v${VERSION}
	hash    = ""

	client  *docker.Client
	oldData *GlobalResponse

	authToken string
	baseURL   string

	timeout   time.Duration
	startTime = time.Now()

	cmd = &cobra.Command{
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
			fmt.Printf("Version: %s - Hash: %s\n", version, hash)
		},
	}
)

func main() {
	setupFlags()
	//*
	h, err := getHash(os.Args[0])
	if err != nil {
		panic(err)
	}
	hash = h
	/**/
	cmd.Long = fmt.Sprintf(longHelp, version, hash)
	cmd.AddCommand(versionCmd, infoCmd, daemonCmd)
	cmd.Execute()
}

func setupFlags() {
	cmd.PersistentFlags().BoolP(VerboseFlag, "v", false, "Turns on verbose logging")
	cmd.PersistentFlags().StringP(EndpointFlag, "e", "unix:///var/run/docker.sock", "Docker endpoint.  Can also set default environment DOCKER_HOST")

	daemonCmd.Flags().DurationVarP(&timeout, TimeoutFlag, "r", 5*time.Minute, "Timeout before force refresh of collected data without event trigger during timeout period")
	daemonCmd.Flags().StringVarP(&authToken, TokenFlag, "t", "", "Firebase authentification token")
	daemonCmd.Flags().StringVarP(&baseURL, BaseURLFlag, "u", "", "Firebase base url")
	//TODO Setup a list modules to load like modules=host,collector,docker ...
	//TODO add flag to force UUID
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
	for id, i := range imgs {
		if len(i.Labels) > 0 { //Reconstruct map without . in key
			tmp := make(map[string]string, len(i.Labels))
			for iid, val := range i.Labels {
				tmp[strings.Replace(iid, ".", "-", -1)] = val
			}
			imgs[id].Labels = tmp
		}
	}

	//Get container
	cnts, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		panic(err)
	}
	for id, c := range cnts {
		if len(c.Labels) > 0 { //Reconstruct map without . in key
			tmp := make(map[string]string, len(c.Labels))
			for vid, val := range c.Labels {
				tmp[strings.Replace(vid, ".", "-", -1)] = val
			}
			cnts[id].Labels = tmp
		}
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
	//Clean of . in key info.RegistryConfig.IndexConfigs
	tmp := make(map[string]*docker.IndexInfo, len(info.RegistryConfig.IndexConfigs))
	for id, conf := range info.RegistryConfig.IndexConfigs {
		tmp[strings.Replace(id, ".", "-", -1)] = conf
	}
	info.RegistryConfig.IndexConfigs = tmp

	//Get networks
	nets, err := client.ListNetworks()
	if err != nil {
		panic(err)
	}
	//Clean . in key of options
	for id, n := range nets {
		if len(n.Options) > 0 { //Reconstruct map without . in key
			tmp := make(map[string]string, len(n.Options))
			for oid, opt := range n.Options {
				tmp[strings.Replace(oid, ".", "-", -1)] = opt
			}
			nets[id].Options = tmp
		}
		if len(n.Labels) > 0 { //Reconstruct map without . in key
			tmp := make(map[string]string, len(n.Labels))
			for lid, val := range n.Labels {
				tmp[strings.Replace(lid, ".", "-", -1)] = val
			}
			nets[id].Labels = tmp
		}
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
		Version:   version,
		StartTime: startTime,
		Hash:      hash,
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

	ints := make([]InterfaceResponse, len(ifaces))
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
	//TODO detect if docket and filter by modules list
	host := getHostData(client)
	u5, err := uuid.NewV5(uuid.NamespaceURL, []byte(host.Name)) //TODO better discriminate maybe add time and save it in /etc/sca/uuid ?
	if err != nil {
		panic(err)
	}
	return &GlobalResponse{
		UUID:      u5.String(),
		Host:      host,
		Collector: getCollectorData(),
		Docker:    getDockerData(client),
	}
}

func sendData(data *GlobalResponse) {
	//TODO update only needed data
	//j, _ := json.Marshal(data)
	//log.Debugln(string(j))

	if oldData == nil {
		log.Debug("Preparing set ...")
		//apiSet(baseURL+"/"+data.UUID, data)
		apiSet(data.UUID, data)
		//Debug
		bytes, _ := json.Marshal(data)
		log.WithFields(log.Fields{
			"data_bytes": len(bytes),
		}).Info("Sending complete messages")
		//log.Debug(data)
		oldData = data //Save state
	} else {
		log.Debug("Preparing update ...")
		if reflect.DeepEqual(oldData, data) {
			log.Debug("Nothing to update data are identical")
			return
		}
		//Debug
		bytes, _ := json.Marshal(data)
		/*
			cleanData := removeDuplicateData("", structs.New(oldData), structs.New(data)) //removeDuplicateData(oldData, data) //cleanData(data) //Remove duplicate
			if err := f.Update(cleanData); err != nil {
				log.Fatal(err)
			}
		*/
		cleanData := sendDeDuplicateData(data.UUID, structs.New(oldData), structs.New(data)) //removeDuplicateData(oldData, data) //cleanData(data) //Remove duplicate

		//Debug
		cleanBytes, _ := json.Marshal(cleanData)
		log.WithFields(log.Fields{
			"data_bytes": len(bytes),
			"send_bytes": len(cleanBytes),
		}).Info("Sending update messages")
		//log.Debug(cleanData)
		oldData = data //Save state of global data
	}
}

func startDaemon(cmd *cobra.Command, args []string) {
	if authToken == "" {
		panic(errors.New("You need to set a auth token"))
	}
	//TODO monitor event and update data
	client := initClient(cmd, args)
	sendData(getData(client))

	c := time.Tick(timeout)
	for now := range c {
		log.Debug("Timeout tick triggered ", now)
		sendData(getData(client))
	}
	//func (c *Client) AddEventListener(listener chan<- *APIEvents) error
}
