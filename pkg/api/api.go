package api

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

//API interface for sca backend
type API struct {
	APIKey       string
	BaseURL      string
	RefreshToken string
	AccessToken  string
	//TODO add queue
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
	return &API{APIKey: apiKey, BaseURL: baseURL, RefreshToken: refreshToken, AccessToken: accessToken}, nil
}

//Send //TODO
func (a *API) Send(data map[string]interface{}) error {
	return nil
}

//SendDeduplicate //TODO
func (a *API) SendDeduplicate(data map[string]interface{}) error {
	return nil
}
