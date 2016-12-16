package main

import (
	"log"

	"github.com/zabawaba99/firego"
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
	f.Auth(authToken)
	defer f.Unauth()
	if err := f.Set(data); err != nil {
		log.Fatal(err)
	}
}
