package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type accessTokenRep struct {
	AccessToken string `json:"access_token"`
	IssuedAt    string `json:"issued_at"`
	ExpiresIn   string `json:"expires_in"`
}

type nswApiTokenManager struct {
	apiKey    string
	apiSecret string

	ApiToken  string
	ExpirySec int
	IssuedAt  time.Time
}

func NewNswApiTokenManager(apiKey string, apiSecret string) *nswApiTokenManager {
	tm := nswApiTokenManager{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}

	return &tm
}

func (tm *nswApiTokenManager) GetOrRenewApiToken() (string, error) {

	reqUrl := "https://api.onegov.nsw.gov.au/oauth/client_credential/accesstoken?grant_type=client_credentials"
	client := http.Client{}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	req.SetBasicAuth(tm.apiKey, tm.apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var rep accessTokenRep
	err = json.Unmarshal(body, &rep)
	if err != nil {
		return "", err
	}

	return rep.AccessToken, nil
}
