package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	firego "gopkg.in/zabawaba99/firego.v1"
)

type refreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}
type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

//TODO detectr fail and replay
//TODO queu FIFO message in order to recover from tiemout and keep message in track

func apiRemove(path string) {
	f := firego.New(baseURL+"/data/"+path, nil)
	f.Auth(authToken)
	defer f.Unauth()
	err := f.Remove()
	switch err := err.(type) {
	case nil:
		// carry on
	default:
		if strings.Compare(err.Error(), "Auth token is expired") == 0 {
			apiGetAuthToken() //TODO get this request in the queue
		} else {
			log.Fatal(err) //TODO handle all errors
		}
		//case firego.EventTypeAuthRevoked:
		//	apiGetAuthToken()
	}
}

func apiSet(path string, data interface{}) {
	f := firego.New(baseURL+"/data/"+path, nil)
	//log.Debug("F set url:" + baseURL + "/" + path)
	f.Auth(authToken)
	//log.Debug("F token set token:" + authToken)
	defer f.Unauth()
	err := f.Set(data)
	switch err := err.(type) {
	case nil:
		// carry on
	default:
		if strings.Compare(err.Error(), "Auth token is expired") == 0 {
			apiGetAuthToken() //TODO get this request in the queue
		} else {
			log.Fatal(err) //TODO handle all errors
		}
		//case firego.EventTypeAuthRevoked:
		//	apiGetAuthToken()
	}
	//log.Debug("F send success")
}

func apiGetAuthToken() {
	log.Debug("Getting new Access Token ... ")
	/*
		var tr *http.Transport
		tr = &http.Transport{
			DisableKeepAlives: true, // https://code.google.com/p/go/issues/detail?id=3514
			Dial: func(network, address string) (net.Conn, error) {
				start := time.Now()
				c, err := net.DialTimeout(network, address, fb.clientTimeout)
				tr.ResponseHeaderTimeout = fb.clientTimeout - time.Since(start)
				return c, err
			},
		}

		client = &http.Client{
			Transport:     tr,
			CheckRedirect: redirectPreserveHeaders,
		}
		req, err := http.NewRequest(method, fb.String(), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
	*/
	/*
		  url := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", apiKey)
			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				log.Fatal("NewRequest: ", err)
				return
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal("Do: ", err)
				return
			}
	*/
	payload, err := json.Marshal(refreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	})
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", apiKey)
	//resp, err := http.PostForm("https://securetoken.googleapis.com/v1/token", url.Values{"key": {apiKey}})
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	defer resp.Body.Close()

	var j refreshResponse
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		log.Fatal(err)
	}
	log.Debug("AccessToken : ", j.AccessToken)
	log.Debug("AccessTokenExpire : ", j.ExpiresIn)

	authToken = j.AccessToken
}
