package main

import (
	log "github.com/sirupsen/logrus"
	firego "gopkg.in/zabawaba99/firego.v1"
)

//TODO detectr fail and replay
//TODO queu FIFO message in order to recover from tiemout and keep message in track

func apiRemove(path string) {
	f := firego.New(baseURL+"data/"+path, nil)
	f.Auth(authToken)
	defer f.Unauth()
	if err := f.Remove(); err != nil {
		log.Fatal(err)
	}
}

func apiSet(path string, data interface{}) {
	f := firego.New(baseURL+"data/"+path, nil)
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
