package api

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/oleiade/lane"
	log "github.com/sirupsen/logrus"
	"github.com/y0ssar1an/q"
	"github.com/zabawaba99/firego"
)

//API interface for sca backend
type API struct {
	APIKey       string
	BaseURL      string
	RefreshToken string
	AccessToken  string
	_data        map[string]interface{}
	_queue       *lane.Queue
	//TODO add queue
}

//QueueItem represente a elemetn of action to send to API
type QueueItem struct {
	Type string
	Data map[string]interface{}
}

//New constructor for API
func New(apiKey, refreshToken, baseURL string) (*API, error) {
	log.WithFields(log.Fields{
		"apiKey":       apiKey,
		"refreshToken": refreshToken,
		"baseURL":      baseURL,
	}).Debug("Init new API")
	//Check params
	if apiKey == "" {
		return nil, errors.New("You need to set a apiKey")
	}
	if refreshToken == "" {
		return nil, errors.New("You need to set a refreshToken")
	}
	if baseURL == "" {
		return nil, errors.New("You need to set a baseURL")
	}
	//Generate frist access token
	accessToken, err := apiGetAuthToken(apiKey, refreshToken)
	if err != nil {
		return nil, err
	}
	return &API{APIKey: apiKey, BaseURL: baseURL, RefreshToken: refreshToken, AccessToken: accessToken, _queue: lane.NewQueue()}, nil
}

func sizeOfJSON(data map[string]interface{}) int {
	//Debug
	bytes, _ := json.Marshal(data)
	return len(bytes)
}

//Send //TODO
func (a *API) Send(data map[string]interface{}) error {
	if a._data == nil { //No data of backend so sending the complet obj
		a.set(data["UUID"].(string), data)
		//TODO -> queue.Enqueue(&QueueItem{Type: "set", Data: data})
		log.WithFields(log.Fields{
			"data_bytes": sizeOfJSON(data), //Debug
		}).Info("Add complete messages to queue")
		a._data = data //Save state
	} else {
		if reflect.DeepEqual(a._data, data) {
			log.Info("Nothing to update data are identical from last send.")
			return nil
		}
		//Debug
		sizeBeforeCleaning := sizeOfJSON(data)
		cleanData := a.sendDeDuplicateData(data["UUID"].(string), a._data, data)
		//TODO at each step -> queue.Enqueue(&QueueItem{Type: "set", Data: data})
		sizeAfterCleaning := sizeOfJSON(cleanData)
		log.WithFields(log.Fields{
			"data_bytes": sizeBeforeCleaning,
			"send_bytes": sizeAfterCleaning,
		}).Info("Sending update messages")
		//log.Debug(cleanData)
		a._data = data //Save state
	}
	//queue.Enqueue(data)
	return nil
}

func (a *API) set(path string, data interface{}) {
	log.WithFields(log.Fields{
		//"api":  a,
		"path": path,
		//"data": data,
	}).Debug("API.set")
	f := firego.New(a.BaseURL+"/data/"+path, nil)
	f.Auth(a.AccessToken)
	defer f.Unauth()
	err := f.Set(data)
	switch err := err.(type) {
	case nil:
		// carry on
	default:
		if strings.Contains(err.Error(), "Auth token is expired") {
			log.WithFields(log.Fields{
				"api": a,
			}).Debug("Auth token is expired -> re-newing AccessToken")
			a.AccessToken, err = apiGetAuthToken(a.APIKey, a.RefreshToken)
			if err != nil {
				log.WithFields(log.Fields{
					"api": a,
				}).Debug("Failed to re-new AccessToken")
			}
			a.set(path, data)
			//TODO get this request in the queue not redo
		} else {
			log.WithFields(log.Fields{
				"api":  a,
				"path": path,
				"data": data,
				"err":  err,
			}).Fatal("Unhandled error in api.set()") //TODO handle all errors
		}
	}
}

func (a *API) remove(path string) {
	log.WithFields(log.Fields{
		//"api":  a,
		"path": path,
	}).Debug("API.remove")
	f := firego.New(a.BaseURL+"/data/"+path, nil)
	f.Auth(a.AccessToken)
	defer f.Unauth()
	err := f.Remove()
	switch err := err.(type) {
	case nil:
		// carry on
	default:
		if strings.Contains(err.Error(), "Auth token is expired") {
			log.WithFields(log.Fields{
				"api": a,
			}).Debug("Auth token is expired -> re-newing AccessToken")
			a.AccessToken, err = apiGetAuthToken(a.APIKey, a.RefreshToken)
			if err != nil {
				log.WithFields(log.Fields{
					"api": a,
				}).Debug("Failed to re-new AccessToken")
			}
			a.remove(path)
			//TODO get this request in the queue not redo
		} else {
			log.WithFields(log.Fields{
				"api":  a,
				"path": path,
				"err":  err,
			}).Fatal("Unhandled error in api.remove()") //TODO handle all errors
		}
	}
}

func (a *API) sendDeDuplicateData(path string, old map[string]interface{}, new map[string]interface{}) map[string]interface{} {
	log.WithFields(log.Fields{
		"path": path,
		//"old":  old,
		//"new": new,
	}).Debug("API.sendDeDuplicateData")
	ret := map[string]interface{}{}

	//Remove old key not in new
	for key := range old {
		if _, ok := new[key]; !ok { //Key not in new we should remove
			a.remove(path + "/" + key)
		}
	}
	//Set new key not in old
	//Parse key in new and old
	for key, newValue := range new {
		if oldvalue, ok := old[key]; !ok { //Key not in old we should set
			a.set(path+"/"+key, newValue)
			ret[key] = newValue //Store in result for stat
		} else { //Key is in new and old -> we recurse or set if final obj differ
			if !reflect.DeepEqual(oldvalue, newValue) { //new differ from old
				if structs.IsStruct(oldvalue) && structs.IsStruct(newValue) { //We have a object -> rescursive
					ret[key] = a.sendDeDuplicateData(path+"/"+key, structs.Map(oldvalue), structs.Map(newValue)) //Store in result for stat
				} else {
					switch t := newValue.(type) {
					//int64
					case int64:
						// t is of type string
						a.set(path+"/"+key, newValue)
						ret[key] = newValue //Store in result for stat
					case string:
						// t is of type string
						a.set(path+"/"+key, newValue)
						ret[key] = newValue //Store in result for stat
					case []string:
						// t is of type array/slice
						a.set(path+"/"+key, newValue)
						ret[key] = newValue //Store in result for stat
						//TODO sort and send only necessary update
					case []interface{}:
						// t is of type array/slice
						a.set(path+"/"+key, newValue)
						ret[key] = newValue //Store in result for stat
						//TODO sort and send only necessary update
					case map[string]interface{}:
						//q.Q(path, newValue)
						ret[key] = a.sendDeDuplicateData(path+"/"+key, oldvalue.(map[string]interface{}), newValue.(map[string]interface{})) //Store in result for stat
					default:
						q.Q(path, newValue)
						log.WithFields(log.Fields{
							"path": path,
							//"old":  old,
							//"new":  new,
							"type": t,
						}).Fatal("Unhandled type in api.sendDeDuplicateData()") //TODO handle all type
					}
				}
			}
		}
	}
	return ret
}

//TODO refactor this
/*
func (a *API) sendDeDuplicateData(path string, sO *structs.Struct, sN *structs.Struct) map[string]interface{} { //map[string]interface{}
	mapR := map[string]interface{}{}
	for _, fO := range sO.Fields() { //Keys in old obj
		_, ok := sN.FieldOk(fO.Name()) //Check if key exist in New
		if !ok {                       //No key found
			log.Debug("Key ", path+"/"+fO.Name(), " from Old is missing in New -> Remove from distant")
			a.remove(path + "/" + fO.Name())
		}
	}
	for _, fN := range sN.Fields() { //Keys in new obj
		fO, ok := sO.FieldOk(fN.Name()) //Check if key exist in New
		if !ok {                        //No key found in old
			mapR[fN.Name()] = fN.Value() //Store in result
			log.Debug("Key ", path+"/"+fN.Name(), " from New is missing in Old -> Set To distant")
			a.set(path+"/"+fN.Name(), fN.Value())
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
				mapR[fN.Name()] = a.sendDeDuplicateData(path+"/"+fN.Name(), structs.New(fO.Value()), structs.New(fN.Value()))
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
						sort.Sort(pkg.ByContainerID(arrN))
						sort.Sort(pkg.ByContainerID(arrO))
						list := make([]map[string]interface{}, len(arrN))
						for i := 0; i < pkg.Min(len(arrN), len(arrO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = a.sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrO[i]), structs.New(arrN[i])) //TODO Report back to mapR
						}

						for i := pkg.Min(len(arrN), len(arrO)); i < len(arrO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							a.remove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := pkg.Min(len(arrN), len(arrO)); i < len(arrN); i++ { //Ajout
							list[i] = structs.Map(arrN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrN[i])
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
						sort.Sort(pkg.ByImageID(arrIN))
						sort.Sort(pkg.ByImageID(arrIO))
						list := make([]map[string]interface{}, len(arrIN))
						for i := 0; i < pkg.Min(len(arrIN), len(arrIO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = a.sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrIO[i]), structs.New(arrIN[i])) //TODO Report back to mapR
						}

						for i := pkg.Min(len(arrIN), len(arrIO)); i < len(arrIO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							a.remove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := pkg.Min(len(arrIN), len(arrIO)); i < len(arrIN); i++ { //Ajout
							list[i] = structs.Map(arrIN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrIN[i])
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
						sort.Sort(pkg.ByVolumeID(arrVN))
						sort.Sort(pkg.ByVolumeID(arrVO))
						list := make([]map[string]interface{}, len(arrVN))
						for i := 0; i < pkg.Min(len(arrVN), len(arrVO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = a.sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrVO[i]), structs.New(arrVN[i])) //TODO Report back to mapR
						}

						for i := pkg.Min(len(arrVN), len(arrVO)); i < len(arrVO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							a.remove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := pkg.Min(len(arrVN), len(arrVO)); i < len(arrVN); i++ { //Ajout
							list[i] = structs.Map(arrVN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrVN[i])
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
						sort.Sort(pkg.ByNetworkID(arrNN))
						sort.Sort(pkg.ByNetworkID(arrNO))
						list := make([]map[string]interface{}, len(arrNN))
						for i := 0; i < pkg.Min(len(arrNN), len(arrNO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a struct")
							list[i] = a.sendDeDuplicateData(path+"/"+fN.Name()+"/"+strconv.Itoa(i), structs.New(arrNO[i]), structs.New(arrNN[i])) //TODO Report back to mapR
						}

						for i := pkg.Min(len(arrNN), len(arrNO)); i < len(arrNO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							a.remove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := pkg.Min(len(arrNN), len(arrNO)); i < len(arrNN); i++ { //Ajout
							list[i] = structs.Map(arrNN[i])
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrNN[i])
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
						for i := 0; i < pkg.Min(len(arrSN), len(arrSO)); i++ { //Compare common
							log.Debug(path+"/"+fN.Name()+"/"+strconv.Itoa(i), " is a string")
							if strings.Compare(arrSO[i], arrSN[i]) != 0 { //Compare string
								list[i] = arrSN[i]
								a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrSN[i]) //Change detected
							}
						}

						for i := pkg.Min(len(arrSN), len(arrSO)); i < len(arrSO); i++ { //Remove
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from Old is missing in New -> Remove from distant")
							a.remove(path + "/" + fN.Name() + "/" + strconv.Itoa(i))
						}

						for i := pkg.Min(len(arrSN), len(arrSO)); i < len(arrSN); i++ { //Ajout
							list[i] = arrSN[i]
							log.Debug("Key ", path+"/"+fN.Name()+"/"+strconv.Itoa(i), " from New is missing in Old -> Set To distant")
							a.set(path+"/"+fN.Name()+"/"+strconv.Itoa(i), arrSN[i])
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
				a.set(path+"/"+fN.Name(), fN.Value())
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
