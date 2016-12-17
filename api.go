package main

import (
	log "github.com/sirupsen/logrus"
	firego "gopkg.in/zabawaba99/firego.v1"
)

func apiRemove(path string) {
	f := firego.New(baseURL+"/"+path, nil)
	f.Auth(authToken)
	defer f.Unauth()
	if err := f.Remove(); err != nil {
		log.Fatal(err)
	}
}

func apiSet(path string, data interface{}) {
	f := firego.New(baseURL+"/"+path, nil)
	//log.Debug("F set url:" + baseURL + "/" + path)
	f.Auth(authToken)
	//log.Debug("F token set token:" + authToken)
	defer f.Unauth()
	//log.Debug("F sending")
	if err := f.Set(data); err != nil {
		//log.Debug("F sending error")
		log.Fatal(err)
	}
	//log.Debug("F send success")
}
