package main

import (
	"encoding/json"
	"fmt"
	"time"

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
			options := map[string]string{
				"version": version,
				"commit":  commit,
			}
			modules := module.GetList(options)
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
	cmd.PersistentFlags().StringP(pkg.EndpointFlag, "e", "unix:///var/run/docker.sock", "Docker endpoint.  Can also set default environment DOCKER_HOST")

	daemonCmd.Flags().DurationVarP(&timeout, pkg.TimeoutFlag, "r", 1*time.Minute, "Timeout before force refresh of collected data without event trigger during timeout period")
	daemonCmd.Flags().StringVarP(&refreshToken, pkg.TokenFlag, "t", "", "Firebase authentification token")
	daemonCmd.Flags().StringVarP(&baseURL, pkg.BaseURLFlag, "u", "", "Firebase base url")
	daemonCmd.Flags().StringVarP(&apiKey, pkg.APIFlag, "k", "", "Firebase api key")
	//TODO Setup a list modules to enable like modules=host,collector,docker ...
	//TODO add flag to force UUID
}

/*
func initDockerClient(cmd *cobra.Command, args []string) *docker.Client {
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
		Version:    version,
		StartTime:  startTime,stResponse {
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

*/

/*
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

*/

func startDaemon(cmd *cobra.Command, args []string) {
	api, err := api.New(apiKey, refreshToken, baseURL) //Init API
	if err != nil {
		log.Fatal("Fail to init API backend", err)
	}
	options := map[string]string{
		"version": version,
		"commit":  commit,
	}
	modules := module.GetList(options)
	api.Send(getData(modules))

	c := time.Tick(timeout)
	for now := range c {
		log.Debug("Timeout tick triggered ", now)
		api.SendDeduplicate(getData(modules))
	}
	//func (c *Client) AddEventListener(listener chan<- *APIEvents) error
}
func getData(modules []module.Module) map[string]interface{} {
	d := make(map[string]interface{})
	for _, m := range modules {
		d[m.ID()] = m.GetData()
	}
	return d
}
