package api

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/oleiade/lane"
	"github.com/sapk/sca/pkg/tools"
	log "github.com/sirupsen/logrus"
	"github.com/zabawaba99/firego"
)

//API interface for sca backend
type API struct {
	APIKey       string
	BaseURL      string
	RefreshToken string
	AccessToken  string
	_data        map[string]interface{}
	_queue       *lane.Deque
	//_update      chan bool
	//TODO add queue
}

//QueueItem represente a elemetn of action to send to API
type QueueItem struct {
	Type string
	Path string
	Data interface{}
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
	return &API{APIKey: apiKey, BaseURL: baseURL, RefreshToken: refreshToken, AccessToken: accessToken, _queue: lane.NewDeque()}, nil
}

//Send data to api with deduction of common value since last update
func (a *API) Send(data map[string]interface{}) error {
	if a._data == nil { //No data of backend so sending the complet obj
		a._data = a.update(data["uuid"].(string), data).(map[string]interface{}) //Save state
		//TODO -> queue.Enqueue(&QueueItem{Type: "set", Data: data})
		log.WithFields(log.Fields{
			"data_bytes": tools.SizeOfJSON(data), //Debug
		}).Info("Add complete messages to queue")
	} else {
		if reflect.DeepEqual(a._data, data) {
			log.Info("Nothing to update data are identical from last send.")
			return nil
		}
		//Debug
		sizeBeforeCleaning := tools.SizeOfJSON(data)
		cleanData, sendedData := a.sendDeDuplicateData(data["uuid"].(string), a._data, data)
		//TODO at each step -> queue.Enqueue(&QueueItem{Type: "set", Data: data})
		sizeAfterCleaning := tools.SizeOfJSON(cleanData)
		log.WithFields(log.Fields{
			"data_bytes": sizeBeforeCleaning,
			"send_bytes": sizeAfterCleaning,
		}).Info("Sending update messages")
		//log.Debug(cleanData)
		a._data = sendedData //Save state
	}

	return a.executeQueue()
}

func (a *API) sendUpdate(updates map[string]interface{}) map[string]interface{} {
	if len(updates) > 0 {
		log.WithFields(log.Fields{
			"size": len(updates),
		}).Debug("sendUpdate") //Send update before set
		a.exectue("UPDT", "", updates)
		return map[string]interface{}{}
	}
	return updates
}
func (a *API) executeQueue() error {
	updates := map[string]interface{}{}
	size := a._queue.Size()

	for i := 0; i < size; i++ {
		value := a._queue.Shift()
		item := value.(*QueueItem)
		switch item.Type {
		case "SET":
			updates = a.sendUpdate(updates)
			a.exectue("SET", item.Path, item.Data)
		case "DEL":
			updates = a.sendUpdate(updates)
			a.exectue("DEL", item.Path, nil)
		case "UPDT":
			updates[item.Path] = item.Data
		default:
			log.WithFields(log.Fields{
				"item": item,
			}).Debug("Unhandled item type in queue.")
		}
	}
	updates = a.sendUpdate(updates)
	return nil
}

func (a *API) set(path string, data interface{}) interface{} {
	log.WithFields(log.Fields{
		"path":      path,
		"queueSize": a._queue.Size(),
	}).Debug("API.set -> add to queue")
	a._queue.Append(&QueueItem{
		Type: "SET",
		Path: path,
		Data: data,
	})
	return data
}

func (a *API) remove(path string) {
	log.WithFields(log.Fields{
		"path":      path,
		"queueSize": a._queue.Size(),
	}).Debug("API.remove -> add to queue")
	a._queue.Append(&QueueItem{
		Type: "DEL",
		Path: path,
	})
}

func (a *API) update(path string, data interface{}) interface{} {
	log.WithFields(log.Fields{
		"path":      path,
		"queueSize": a._queue.Size(),
	}).Debug("API.update -> add to queue")
	a._queue.Append(&QueueItem{
		Type: "UPDT",
		Path: path,
		Data: data,
	})
	return data
}

func (a *API) exectue(method string, path string, data interface{}) {
	log.WithFields(log.Fields{
		//"api": a,
		"method": method,
		"path":   path,
		//"keys":   data,
		//"data":   data,
	}).Debug("API.execute")
	f := firego.New(a.BaseURL+"/data/"+path, nil)
	f.Auth(a.AccessToken)
	defer f.Unauth()

	var err error
	switch method {
	case "SET":
		err = f.Set(data)
	case "DEL":
		err = f.Remove()
	case "UPDT":
		err = f.Update(data)
	}

	//Handleling errors
	switch err := err.(type) {
	case nil:
		return
	default:
		if strings.Contains(err.Error(), "Auth token is expired") {
			log.WithFields(log.Fields{
				"api.AccessToken": a.AccessToken,
			}).Debug("Auth token is expired -> re-newing AccessToken")
			a.AccessToken, err = apiGetAuthToken(a.APIKey, a.RefreshToken)
			if err != nil {
				log.WithFields(log.Fields{
					"api": a,
				}).Debug("Failed to re-new AccessToken")
			}
			a.exectue(method, path, data) //Redo
			return
		}
		if strings.Contains(err.Error(), "Internal server error.") {
			log.WithFields(log.Fields{
				"api.AccessToken": a.AccessToken,
				"method":          method,
				"path":            path,
				"data":            data,
				"err":             err,
			}).Warning("API respond with : Internal server error. -> skipping update")
			//TODO force set of _data to do not have any inconsistency
		} //else {
		log.WithFields(log.Fields{
			//"api":  a,
			"method": method,
			"path":   path,
			"data":   data,
			"err":    err,
		}).Fatal("Unhandled error in api.execut()") //TODO handle all errors
		return
	}

}

func (a *API) sendDeDuplicateData(path string, old map[string]interface{}, new map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	log.WithFields(log.Fields{
		"path": path,
		//"old":  old,
		//"new": new,
	}).Debug("API.sendDeDuplicateData")
	ret := map[string]interface{}{}
	realRet := map[string]interface{}{}

	//Remove old key not in new
	for key := range old {
		if _, ok := new[key]; !ok { //Key not in new we should remove
			a.remove(path + "/" + key)
		}
	}
	//Set new key not in old
	//Parse key in new and old
	for key, newValue := range new {
		if oldValue, ok := old[key]; !ok { //Key not in old we should set
			ret[key] = a.update(path+"/"+key, newValue)
			realRet[key] = ret[key]
		} else { //Key is in new and old -> we recurse or set if final obj differ
			if !reflect.DeepEqual(oldValue, newValue) { //new differ from old
				if structs.IsStruct(oldValue) && structs.IsStruct(newValue) { //We have a object -> rescursive
					ret[key], realRet[key] = a.sendDeDuplicateData(path+"/"+key, structs.Map(oldValue), structs.Map(newValue)) //Store in result for stat
				} else {
					switch newValue.(type) {
					case bool, int, int32, int64, uint, uint32, uint64, float32, float64, string, []string: //Simple array are ordered so if there a diff we update
						ret[key] = a.update(path+"/"+key, newValue)
						realRet[key] = ret[key]
					case [][2]string:
						// t is of type array/slice
						ret[key] = a.update(path+"/"+key, newValue)
						realRet[key] = ret[key]
						//TODO send only necessary update
					case []interface{}:
						// t is of type array/slice
						newValueArr := newValue.([]interface{})
						oldValueArr := oldValue.([]interface{})
						commonMin := tools.Min(len(newValueArr), len(oldValueArr))
						list := make([]interface{}, len(newValueArr))
						listR := make([]interface{}, len(newValueArr))
						for i := 0; i < commonMin; i++ { //Compare common
							if structs.IsStruct(oldValueArr[i]) && structs.IsStruct(newValueArr[i]) { //We have a object -> rescursive
								list[i], listR[i] = a.sendDeDuplicateData(path+"/"+key+"/"+strconv.Itoa(i), structs.Map(oldValueArr[i]), structs.Map(newValueArr[i]))
							} else {
								switch newValueArr[i].(type) {
								case map[string]interface{}: //Allready map
									list[i], listR[i] = a.sendDeDuplicateData(path+"/"+key+"/"+strconv.Itoa(i), oldValueArr[i].(map[string]interface{}), newValueArr[i].(map[string]interface{}))
								default: //Force update
									log.WithFields(log.Fields{
										//"api":  a,
										"path": path + "/" + key + "/" + strconv.Itoa(i),
										"data": newValueArr[i],
									}).Debug("Force api.update() on data since it seems to not be a struct")
									list[i] = a.update(path+"/"+key+"/"+strconv.Itoa(i), newValueArr[i])
									listR[i] = list[i]
								}
							}
						}
						for i := commonMin; i < len(oldValueArr); i++ { //Remove
							a.remove(path + "/" + key + "/" + strconv.Itoa(i))
						}
						for i := commonMin; i < len(newValueArr); i++ { //Add
							list[i] = a.update(path+"/"+key+"/"+strconv.Itoa(i), newValueArr[i])
							listR[i] = list[i]
							/*
								log.WithFields(log.Fields{
									//"api":  a,
									"path": path + "/" + key + "/" + strconv.Itoa(i),
									"data": newValueArr[i],
								}).Debug("Force api.set() on data since it seems over old array size")
							*/
						}
						ret[key] = list
						realRet[key] = listR
					case map[string]interface{}:
						//q.Q(path, newValue)
						ret[key], realRet[key] = a.sendDeDuplicateData(path+"/"+key, oldValue.(map[string]interface{}), newValue.(map[string]interface{})) //Store in result for stat
					default:
						//q.Q(path, newValue)
						log.WithFields(log.Fields{
							"path": path,
							//"old":  old,
							//"new":  new,
						}).Warn("Unhandled type in api.sendDeDuplicateData() falling back to coping all object") //TODO handle all type
						ret[key] = a.update(path+"/"+key, newValue)
						realRet[key] = ret[key]
					}
				}
			} else {
				//Old and New are equal we update real global data object
				realRet[key] = newValue
			}
		}
	}
	return ret, realRet
}
