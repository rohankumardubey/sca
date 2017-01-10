package docker

//TODO monitor event and update data
import (
	"sort"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/sapk/sca/pkg"
	log "github.com/sirupsen/logrus"
)

const id = "Docker"

//Module retrieve information form executing sca
type Module struct {
	Endpoint string
	Client   *docker.Client
}

//Response describe docker informations
type Response struct {
	Info       *docker.DockerInfo     `json:"Info,omitempty"`
	Containers []docker.APIContainers `json:"Containers,omitempty"`
	Images     []docker.APIImages     `json:"Images,omitempty"`
	Volumes    []docker.Volume        `json:"Volumes,omitempty"`
	Networks   []docker.Network       `json:"Networks,omitempty"`
}

//New constructor for Module
func New(options map[string]string) *Module {
	log.WithFields(log.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new Module")

	client, err := docker.NewClient(options["docker.endpoint"])
	if err != nil {
		log.WithFields(log.Fields{
			"client": client,
			"err":    err,
		}).Warn("Failed to create docker client")
		//return nil
	}
	return &Module{Endpoint: options["docker.endpoint"], Client: client}
}

//ID //TODO
func (d *Module) ID() string {
	return id
}

//GetData //TODO
func (d *Module) GetData() interface{} {

	return Response{
		Info:       d.getInfo(),
		Containers: d.getContainers(),
		Networks:   d.getNetworks(),
		Volumes:    d.getVolumes(),
		Images:     d.getImages(),
	}
}
func (d *Module) getInfo() *docker.DockerInfo {
	//Get server info
	info, err := d.Client.Info()
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"info":   info,
			"client": d.Client,
		}).Warn("Failed to get docker host info")
		return nil
	}
	//Clean of . in key info.RegistryConfig.IndexConfigs
	tmp := make(map[string]*docker.IndexInfo, len(info.RegistryConfig.IndexConfigs))
	for id, conf := range info.RegistryConfig.IndexConfigs {
		tmp[strings.Replace(id, ".", "-", -1)] = conf
	}
	info.RegistryConfig.IndexConfigs = tmp

	//Sort Docker/Info/Swarm/RemoteManagers/X to ease optimisation on sync
	sort.Sort(pkg.ByPeer(info.Swarm.RemoteManagers))
	return info
}

func (d *Module) getImages() []docker.APIImages {
	//Get images
	imgs, err := d.Client.ListImages(docker.ListImagesOptions{All: true})
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
	sort.Sort(pkg.ByIID(imgs))
	return imgs
}

func (d *Module) getNetworks() []docker.Network {
	//Get networks
	nets, err := d.Client.ListNetworks()
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"nets":   nets,
			"client": d.Client,
		}).Warn("Failed to get docker network list")
		return nil
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
	sort.Sort(pkg.ByNID(nets))
	return nets
}

func (d *Module) getContainers() []docker.APIContainers {
	//Get container
	cnts, err := d.Client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"cnts":   cnts,
			"client": d.Client,
		}).Warn("Failed to get docker container list")
		return nil
	}
	for id, c := range cnts {
		if len(c.Labels) > 0 { //Reconstruct map without . in key
			tmp := make(map[string]string, len(c.Labels))
			for vid, val := range c.Labels {
				tmp[strings.Replace(vid, ".", "-", -1)] = val
			}
			cnts[id].Labels = tmp
		}
		//Sort Docker/Containers/X/Mounts/X to ease optimisation on sync
		sort.Sort(pkg.ByMount(c.Mounts))
		//Sort Docker/Containers/X/Ports/X to ease optimisation on sync
		sort.Sort(pkg.ByPort(c.Ports))
	}
	sort.Sort(pkg.ByCID(cnts))
	return cnts
}

func (d *Module) getVolumes() []docker.Volume {
	//Get volumes
	vols, err := d.Client.ListVolumes(docker.ListVolumesOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"vols":   vols,
			"client": d.Client,
		}).Warn("Failed to get docker volume list")
		return nil
	}
	sort.Sort(pkg.ByVName(vols))
	return vols
}
