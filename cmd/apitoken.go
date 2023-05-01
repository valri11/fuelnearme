package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

type accessTokenRep struct {
	AccessToken string `json:"access_token"`
	IssuedAt    string `json:"issued_at"`
	ExpiresIn   string `json:"expires_in"`
}

type nswApiTokenManager struct {
	ApiToken  string
	ExpirySec int
	IssuedAt  time.Time

	tokenSource oauth2.TokenSource
}

type tokenSource struct {
	tokenURL  string
	apiKey    string
	apiSecret string

	client *http.Client
}

func NewFuelApiTokenSource(tokenUrl string, apiKey string, apiSecret string) (*tokenSource, error) {

	client := http.Client{}

	ts := tokenSource{
		tokenURL:  tokenUrl,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    &client,
	}
	return &ts, nil
}

func (ts *tokenSource) Token() (*oauth2.Token, error) {
	req, err := http.NewRequest("GET", ts.tokenURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	req.SetBasicAuth(ts.apiKey, ts.apiSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rep accessTokenRep
	err = json.Unmarshal(body, &rep)
	if err != nil {
		return nil, err
	}

	issuedAtMsec, err := strconv.Atoi(rep.IssuedAt)
	if err != nil {
		return nil, err
	}
	expiresIn, err := strconv.Atoi(rep.ExpiresIn)
	if err != nil {
		return nil, err
	}

	expiry := time.UnixMilli(int64(issuedAtMsec))
	expiry.Add(time.Duration(expiresIn) * time.Second)

	tk := oauth2.Token{
		AccessToken: rep.AccessToken,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}

	return &tk, nil
}

func NewNswApiTokenManager(apiKey string, apiSecret string) (*nswApiTokenManager, error) {
	tokenUrl := "https://api.onegov.nsw.gov.au/oauth/client_credential/accesstoken?grant_type=client_credentials"
	ts, err := NewFuelApiTokenSource(tokenUrl, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	reusableSource := oauth2.ReuseTokenSource(nil, ts)

	tm := nswApiTokenManager{
		tokenSource: reusableSource,
	}

	return &tm, nil
}

func (tm *nswApiTokenManager) GetOrRenewApiToken() (string, error) {

	tk, err := tm.tokenSource.Token()
	if err != nil {
		return "", err
	}

	return tk.AccessToken, nil
}
