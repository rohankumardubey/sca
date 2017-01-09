package pkg

import (
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//Min tools minimum of int
func Min(A, B int) int {
	min := A
	if A > B {
		min = B
	}
	return min
}

//SortedKeys tools sort map[string]
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

//TypeOrEnv parse cmd to file with env vars
func TypeOrEnv(cmd *cobra.Command, flag, envname string) string {
	val, _ := cmd.Flags().GetString(flag)
	if val == "" {
		val = os.Getenv(envname)
	}
	return val
}

//SetupLogger parse cmd for log level
func SetupLogger(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool(VerboseFlag); verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

/*
//GetHash return hash of a file
func GetHash(filePath string) (result string, err error) {
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
*/
/*
//ByContainerID sort class
type ByContainerID []docker.APIContainers

func (a ByContainerID) Len() int           { return len(a) }
func (a ByContainerID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByContainerID) Less(i, j int) bool { return a[i].ID < a[j].ID }

//ByImageID sort class
type ByImageID []docker.APIImages

func (a ByImageID) Len() int           { return len(a) }
func (a ByImageID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByImageID) Less(i, j int) bool { return a[i].ID < a[j].ID }

//ByVolumeID sort class
type ByVolumeID []docker.Volume

func (a ByVolumeID) Len() int           { return len(a) }
func (a ByVolumeID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVolumeID) Less(i, j int) bool { return a[i].Name < a[j].Name }

//ByNetworkID sort class
type ByNetworkID []docker.Network

func (a ByNetworkID) Len() int           { return len(a) }
func (a ByNetworkID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNetworkID) Less(i, j int) bool { return a[i].ID < a[j].ID }
*/
/*
func sendDeDuplicateData(path string, sO *structs.Struct, sN *structs.Struct) map[string]interface{} { //map[string]interface{}
	//sR := structs.Struct{}
	mapR := map[string]interface{}{}
	//keysO := sortedKeys(mapO)
	//keysN := sortedKeys(mapN)
	for _, fO := range sO.Fields() { //Keys in old obj
		//Debug log.Debug("Parsing key ", path+"/"+fO.Name(), " in Old")
		//vO := mapO[kO]     //Valuer (interface)
		//vN, ok := mapN[kO] //Check if key exist in New
		_, ok := sN.FieldOk(fO.Name()) //Check if key exist in New
		if !ok {                       //No key found
			log.Debug("Key ", path+"/"+fO.Name(), " from Old is missing in New -> Remove from distant")
			apiRemove(path + "/" + fO.Name())
			//return sN.Map() //Send complete to update all object/array
		}
	}
	for _, fN := range sN.Fields() { //Keys in new obj
		//Debug log.Debug("Parsing key ", path+"/"+fN.Name(), " in New")
		//vN := fN.Value()                //Valuer (interface)
		fO, ok := sO.FieldOk(fN.Name()) //Check if key exist in New
		if !ok {                        //No key found in old
			mapR[fN.Name()] = fN.Value() //Store in result
			log.Debug("Key ", path+"/"+fN.Name(), " from New is missing in Old -> Set To distant")
			apiSet(path+"/"+fN.Name(), fN.Value())
			continue
		}
		if fN.IsExported() && fO.IsExported() && !reflect.DeepEqual(fO.Value(), fN.Value()) {
			log.Debug(path+"/"+fN.Name(), " seems to be different")
			log.Debug(path+"/"+fN.Name(), " old kind ", fO.Kind(), " value ", fO.Value())
			log.Debug(path+"/"+fN.Name(), " new kind ", fN.Kind(), " value ", fN.Value())
			//*
			if structs.IsStruct(fO.Value()) && structs.IsStruct(fN.Value()) {
				log.Debug(path+"/"+fN.Name(), " is a struct")
				//mapR[fN.Name()] = removeDuplicateData(path+"/"+fN.Name(), structs.New(fO.Value()), structs.New(fN.Value())) //Store in result of parsing recursive
				mapR[fN.Name()] = sendDeDuplicateData(path+"/"+fN.Name(), structs.New(fO.Value()), structs.New(fN.Value()))
			} else {
				log.Debug(path+"/"+fN.Name(), " is not a struct")
				if fN.Kind() == reflect.Slice {
					log.Debug("Is array !")
					//TODO []net.Addr []InterfaceResponse

					// Catch APIContainers
					arrN, ok1 := fN.Value().([]docker.APIContainers)
					arrO, ok2 := fO.Value().([]docker.APIContainers)
					if ok1 && ok2 {
						log.Debug("Is container array !")
						sort.Sort(ByContainerID(arrN))
						sort.Sort(ByContainerID(arrO))
						list := make([]map[string]interface{}, len(arrN))
						for i := 0; i < min(len(arrN), len(arrO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrO[i]), structs.New(arrN[i])) //TODO Report back to mapR
						}

						for i := min(len(arrN), len(arrO)); i < len(arrO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							apiRemove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := min(len(arrN), len(arrO)); i < len(arrN); i++ { //Ajout
							list[i] = structs.Map(arrN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrN[i])
						}
						mapR[fN.Name()] = list
						continue
					} else {
						log.Debug("Is not a APIContainers array !")
					}

					// Catch APIImages
					arrIN, ok1 := fN.Value().([]docker.APIImages)
					arrIO, ok2 := fO.Value().([]docker.APIImages)
					if ok1 && ok2 {
						log.Debug("Is image array !")
						sort.Sort(ByImageID(arrIN))
						sort.Sort(ByImageID(arrIO))
						list := make([]map[string]interface{}, len(arrIN))
						for i := 0; i < min(len(arrIN), len(arrIO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrIO[i]), structs.New(arrIN[i])) //TODO Report back to mapR
						}

						for i := min(len(arrIN), len(arrIO)); i < len(arrIO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							apiRemove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := min(len(arrIN), len(arrIO)); i < len(arrIN); i++ { //Ajout
							list[i] = structs.Map(arrIN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrIN[i])
						}
						mapR[fN.Name()] = list
						continue
					} else {
						log.Debug("Is not a APIImages array !")
					}

					// Catch Volume
					arrVN, ok1 := fN.Value().([]docker.Volume)
					arrVO, ok2 := fO.Value().([]docker.Volume)
					if ok1 && ok2 {
						log.Debug("Is Volume array !")
						sort.Sort(ByVolumeID(arrVN))
						sort.Sort(ByVolumeID(arrVO))
						list := make([]map[string]interface{}, len(arrVN))
						for i := 0; i < min(len(arrVN), len(arrVO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrVO[i]), structs.New(arrVN[i])) //TODO Report back to mapR
						}

						for i := min(len(arrVN), len(arrVO)); i < len(arrVO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							apiRemove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := min(len(arrVN), len(arrVO)); i < len(arrVN); i++ { //Ajout
							list[i] = structs.Map(arrVN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrVN[i])
						}
						mapR[fN.Name()] = list
						continue
					} else {
						log.Debug("Is not a Volume array !")
					}

					// Catch Network
					arrNN, ok1 := fN.Value().([]docker.Network)
					arrNO, ok2 := fO.Value().([]docker.Network)
					if ok1 && ok2 {
						log.Debug("Is Network array !")
						sort.Sort(ByNetworkID(arrNN))
						sort.Sort(ByNetworkID(arrNO))
						list := make([]map[string]interface{}, len(arrNN))
						for i := 0; i < min(len(arrNN), len(arrNO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrNO[i]), structs.New(arrNN[i])) //TODO Report back to mapR
						}

						for i := min(len(arrNN), len(arrNO)); i < len(arrNO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							apiRemove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := min(len(arrNN), len(arrNO)); i < len(arrNN); i++ { //Ajout
							list[i] = structs.Map(arrNN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrNN[i])
						}
						mapR[fN.Name()] = list
						continue
					} else {
						log.Debug("Is not a Network array !")
					}

					// Catch Strings
					arrSN, ok1 := fN.Value().([]string)
					arrSO, ok2 := fO.Value().([]string)
					if ok1 && ok2 {
						log.Debug("Is string array !")
						sort.Strings(arrSN)
						sort.Strings(arrSO)
						list := make([]string, len(arrSN))
						for i := 0; i < min(len(arrSN), len(arrSO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a string")
							if strings.Compare(arrSO[i], arrSN[i]) != 0 { //Compare string
								list[i] = arrSN[i]
								apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrSN[i]) //Change detected
							}
						}

						for i := min(len(arrSN), len(arrSO)); i < len(arrSO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							apiRemove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := min(len(arrSN), len(arrSO)); i < len(arrSN); i++ { //Ajout
							list[i] = arrSN[i]
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							apiSet(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrSN[i])
						}
						mapR[fN.Name()] = list
						continue
					} else {
						log.Debug("Is not a string array !")
					}
				}
				//else {
				mapR[fN.Name()] = fN.Value()
				log.Debug("Key ", path+"/"+fN.Name(), " from New is not a struct and differ from Old -> Set To distant")
				apiSet(path+"/"+fN.Name(), fN.Value())
				//}
			}
		} else {
			//Debug log.Debug(path+"/"+fN.Name(), " seems to be identical")
		}
	}

	log.Debug(mapR)
	return mapR
}
*/
