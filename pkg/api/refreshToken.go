package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
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

//apiGetAuthToken get a access token form a refresh token and api key
func apiGetAuthToken(apiKey string, refreshToken string) (string, error) {
	log.Info("Getting new Access Token ... ")
	payload, err := json.Marshal(refreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", apiKey)
	//resp, err := http.PostForm("https://securetoken.googleapis.com/v1/token", url.Values{"key": {apiKey}})
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var js refreshResponse
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&js); err != nil {
		return "", err
	}
	log.Debug("AccessToken : ", js.AccessToken)
	log.Debug("AccessTokenExpire : ", js.ExpiresIn)

	return js.AccessToken, nil
}
