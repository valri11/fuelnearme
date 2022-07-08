package geocode

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type RevGeocodeReq struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ApiKey string  `json:"apiKey"`
}

type GeocodeAddress struct {
	StreetName    string `json:"streetName"`
	Surburb       string `json:"suburb"`
	State         string `json:"state"`
	PostCode      string `json:"postCode"`
	StreetAddress string `json:streetAddress"`
}

type RevGeocodeRep struct {
	Type    string         `json:"type"`
	Address GeocodeAddress `json:"result"`
}

func AddressFromLatLon(lat float64, lon float64, apiKey string) (*GeocodeAddress, error) {
	reqUrl := "https://mappify.io/api/rpc/coordinates/reversegeocode"

	r := RevGeocodeReq{
		Lat:    lat,
		Lon:    lon,
		ApiKey: apiKey,
	}

	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)

	client := http.Client{}

	req, err := http.NewRequest("POST", reqUrl, reader)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rep RevGeocodeRep
	err = json.Unmarshal(body, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.Address, nil
}
