package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"reflect"
	"sort"

	"github.com/fatih/structs"
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

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func removeDuplicateData(path string, sO *structs.Struct, sN *structs.Struct) map[string]interface{} {
	//sR := structs.Struct{}
	mapR := map[string]interface{}{}
	//keysO := sortedKeys(mapO)
	//keysN := sortedKeys(mapN)
	for _, fO := range sO.Fields() { //Keys in old obj
		log.Debug("Parsing key ", path+"/"+fO.Name(), " in Old")
		//vO := mapO[kO]     //Valuer (interface)
		//vN, ok := mapN[kO] //Check if key exist in New
		_, ok := sN.FieldOk(fO.Name()) //Check if key exist in New
		if !ok {                       //No key found
			log.Debug("Key ", path+"/"+fO.Name(), " from Old is missing in New")
			return sN.Map() //Send complete to update all object/array
		}
	}
	for _, fN := range sN.Fields() { //Keys in new obj
		log.Debug("Parsing key ", path+"/"+fN.Name(), " in New")
		//vN := fN.Value()                //Valuer (interface)
		fO, ok := sO.FieldOk(fN.Name()) //Check if key exist in New
		if !ok {                        //No key found in old
			mapR[fN.Name()] = fN.Value() //Store in result
			continue
		}
		if !reflect.DeepEqual(fO.Value(), fN.Value()) {
			/*
				switch path + "/" + fN.Name() {
				case "/Docker":
					if tmp, ok := (fN.Value()).(*DockerResponse); ok {
						//vN := tmp
						log.Debug(tmp)
					} else {
						panic("Error of format conversion")
					}
					if tmp, ok := (fO.Value()).(*DockerResponse); ok {
						//vO := tmp
						log.Debug(tmp)
					} else {
						panic("Error of format conversion")
					}
					break
				}
				//*/
			log.Debug(path+"/"+fN.Name(), " seems to be different")
			log.Debug(path+"/"+fN.Name(), " old kind ", fO.Kind(), " value ", fO.Value())
			log.Debug(path+"/"+fN.Name(), " new kind ", fN.Kind(), " value ", fN.Value())
			//*
			if structs.IsStruct(fO.Value()) && structs.IsStruct(fN.Value()) {
				log.Debug(path+"/"+fN.Name(), " is a struct")
				mapR[fN.Name()] = removeDuplicateData(path+"/"+fN.Name(), structs.New(fO.Value()), structs.New(fN.Value())) //Store in result of parsing recursive
			} else {
				log.Debug(path+"/"+fN.Name(), " is not a struct")
				mapR[fN.Name()] = fN.Value()
			}
			//*/
			//TODO maybe order array ?
		} else {
			log.Debug(path+"/"+fN.Name(), " seems to be identical")
		}
	}

	log.Debug(mapR)
	return mapR
}

/*
func removeDuplicateData(old *GlobalResponse, new *GlobalResponse) map[string]interface{} {
	return removeDuplicateData(structs.Map(old), structs.Map(new)) //Empty for now
}
*/

/*
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
*/
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
