package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sort"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//ByContainerID sort class
type ByContainerID []docker.APIContainers

func (a ByContainerID) Len() int           { return len(a) }
func (a ByContainerID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByContainerID) Less(i, j int) bool { return a[i].ID < a[j].ID }

func getHash(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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

func cleanData(data *GlobalResponse) *GlobalResponse {
	//TODO should be done a JSON level
	//Global oldData
	//TODO Make it recursive
	ret := &GlobalResponse{}
	if !reflect.DeepEqual(oldData.Host, data.Host) {
		ret.Host = data.Host
	} else {
		log.Debug("Skipping Host part")
	}
	if !reflect.DeepEqual(oldData.Collector, data.Collector) {
		ret.Collector = data.Collector
	} else {
		log.Debug("Skipping Collector part")
	}

	if !reflect.DeepEqual(oldData.Docker, data.Docker) {
		ret.Docker = &DockerResponse{} //Build empty shell
	}
	//docker = data.Docker
	//docker := &DockerResponse{}
	sort.Sort(ByContainerID(*data.Docker.Containers))
	sort.Sort(ByContainerID(*oldData.Docker.Containers)) //TODO should not be necessary ?
	if !reflect.DeepEqual(oldData.Docker.Containers, data.Docker.Containers) {
		ret.Docker.Containers = data.Docker.Containers
	} else {
		log.Debug("Skipping Docker.Containers part")
	}

	if !reflect.DeepEqual(oldData.Docker.Images, data.Docker.Images) {
		ret.Docker.Images = data.Docker.Images
	} else {
		log.Debug("Skipping Docker.Images part")
	}

	if !reflect.DeepEqual(oldData.Docker.Info, data.Docker.Info) {
		//ret.Docker.Info = &docker.DockerInfo{}
		info := &docker.DockerInfo{}
		//Getting changin values
		if !reflect.DeepEqual(oldData.Docker.Info.NGoroutines, data.Docker.Info.NGoroutines) {
			oldData.Docker.Info.NGoroutines = data.Docker.Info.NGoroutines
			info.NGoroutines = data.Docker.Info.NGoroutines
		}
		if !reflect.DeepEqual(oldData.Docker.Info.NFd, data.Docker.Info.NFd) {
			oldData.Docker.Info.NFd = data.Docker.Info.NFd
			info.NFd = data.Docker.Info.NFd
		}
		if !reflect.DeepEqual(oldData.Docker.Info.NEventsListener, data.Docker.Info.NEventsListener) {
			oldData.Docker.Info.NEventsListener = data.Docker.Info.NEventsListener
			info.NEventsListener = data.Docker.Info.NEventsListener
		}

		sort.Strings(data.Docker.Info.Plugins.Network)
		sort.Strings(oldData.Docker.Info.Plugins.Network) //TODO should not be necessary ?
		sort.Strings(data.Docker.Info.Plugins.Volume)
		sort.Strings(oldData.Docker.Info.Plugins.Volume) //TODO should not be necessary ?
		if !reflect.DeepEqual(oldData.Docker.Info.Plugins, data.Docker.Info.Plugins) {
			jo, _ := json.Marshal(oldData.Docker.Info.Plugins)
			jn, _ := json.Marshal(data.Docker.Info.Plugins)
			if bytes.Compare(jo, jn) != 0 { //Not same json rep
				oldData.Docker.Info.Plugins = data.Docker.Info.Plugins
				info.Plugins = data.Docker.Info.Plugins
				log.Debug(string(jo))
				log.Debug(string(jn))
			}
		}

		if !reflect.DeepEqual(oldData.Docker.Info.Swarm, data.Docker.Info.Swarm) {
			oldData.Docker.Info.Swarm = data.Docker.Info.Swarm
			info.Swarm = data.Docker.Info.Swarm
		}

		if !reflect.DeepEqual(oldData.Docker.Info.SystemTime, data.Docker.Info.SystemTime) {
			oldData.Docker.Info.SystemTime = data.Docker.Info.SystemTime
			info.SystemTime = data.Docker.Info.SystemTime
		}

		if !reflect.DeepEqual(oldData.Docker.Info, data.Docker.Info) {
			//If still not the same
			log.Debug("Docker.Info still not the same after checking common changing var")
			log.Debug(oldData.Docker.Info)
			log.Debug(data.Docker.Info)
			log.Debug(ret.Docker.Info)
			ret.Docker.Info = data.Docker.Info
		} else if !reflect.DeepEqual(info, &docker.DockerInfo{}) {
			ret.Docker.Info = info
		}
	} else {
		log.Debug("Skipping Docker.Info part")
	}

	if !reflect.DeepEqual(oldData.Docker.Networks, data.Docker.Networks) {
		ret.Docker.Networks = data.Docker.Networks
	} else {
		log.Debug("Skipping Docker.Networks part")
	}

	if !reflect.DeepEqual(oldData.Docker.Volumes, data.Docker.Volumes) {
		ret.Docker.Volumes = data.Docker.Volumes
	} else {
		log.Debug("Skipping Docker.Volumes part")
	}
	log.Debug("Host: ", ret.Host)
	log.Debug("Collector: ", ret.Collector)
	log.Debug("Docker.Containers: ", ret.Docker.Containers)
	log.Debug("Docker.Images: ", ret.Docker.Images)
	log.Debug("Docker.Info: ", ret.Docker.Info)
	log.Debug("Docker.Networks: ", ret.Docker.Networks)
	log.Debug("Docker.Volumes: ", ret.Docker.Volumes)
	return ret
}

/*
//Quick and dirty
func convertData(data *GlobalResponse) (map[string]string, error) {
	buffer, _ := json.Marshal(data)
	var m map[string]stringbytes
	if err := json.Unmarshal(buffer, &m); err != nil {
		return nil, err
	}
	return m, nil
}
*/
