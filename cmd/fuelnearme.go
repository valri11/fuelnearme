/*
Copyright © 2022 Val Gridnev <valer.gr@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valri11/fuelnearme/config"
	"github.com/valri11/fuelnearme/geocode"
)

var reqCounter int

type FuelNearmeReq struct {
	FuelType      string   `json:"fueltype"`
	Brand         []string `json:"brand"`
	NamedLocation string   `json:"namedlocation"`
	Lat           float64  `json:"latitude,string"`
	Lon           float64  `json:"longitude,string"`
	SortBy        string   `json:"sortby"`
	SortAscending string   `json:"sortascending"`
	Radius        float64  `json:"radius,string"`
}

func NewFuelNearmeReq() FuelNearmeReq {
	r := FuelNearmeReq{
		Brand:         make([]string, 0),
		SortBy:        "Price",
		SortAscending: "true",
		Radius:        5,
	}
	return r
}

// fuelnearmeCmd represents the fuelnearme command
var fuelnearmeCmd = &cobra.Command{
	Use:   "fuelnearme",
	Short: "Returns nearby fuel prices in GeoJSON format",
	Long:  `Returns nearby fuel prices in GeoJSON format`,
	Run: func(cmd *cobra.Command, args []string) {

		var cfg config.Configuration
		err := viper.Unmarshal(&cfg)
		if err != nil {
			log.Fatalf("error unmarshal config: %v", err)
			return
		}

		lat, err := cmd.Flags().GetFloat64("lat")
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}
		lon, err := cmd.Flags().GetFloat64("lon")
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}
		radius, err := cmd.Flags().GetFloat64("radius")
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		tm := NewNswApiTokenManager(cfg.NswFuelApi.ApiKey, cfg.NswFuelApi.ApiSecret)

		addr, err := geocode.AddressFromLatLon(lat, lon, cfg.Mapify.ApiKey)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		r := NewFuelNearmeReq()

		r.FuelType = "E10"
		r.Lat = lat
		r.Lon = lon
		r.Radius = radius
		r.NamedLocation = addr.PostCode

		data, err := json.Marshal(r)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}
		reader := bytes.NewReader(data)

		reqUrl := "https://api.onegov.nsw.gov.au/FuelPriceCheck/v2/fuel/prices/nearby"

		client := http.Client{}

		req, err := http.NewRequest("POST", reqUrl, reader)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		apiToken, err := tm.GetOrRenewApiToken()
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		reqTimestamp := time.Now().Format("02/01/2006 15:04:05")

		req.Header = http.Header{
			"Content-Type":     {"application/json"},
			"Authorization":    {fmt.Sprintf("Bearer %s", apiToken)},
			"apikey":           {cfg.NswFuelApi.ApiKey},
			"transactionid":    {fmt.Sprintf("tr_%5d", reqCounter)},
			"requesttimestamp": {reqTimestamp},
		}
		reqCounter++

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		fc, err := fuelNearmeDataToFeatureCollection(body)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}

		out, err := json.Marshal(fc)
		if err != nil {
			log.Fatalf("ERR: %v", err)
			return
		}
		fmt.Printf("%s\n", string(out))
	},
}

func init() {
	rootCmd.AddCommand(fuelnearmeCmd)

	fuelnearmeCmd.Flags().Float64("lat", -33.856159, "latitude")
	fuelnearmeCmd.Flags().Float64("lon", 151.215256, "longitude")
	fuelnearmeCmd.Flags().Float64("radius", 5.0, "radius")
}

func fuelNearmeDataToFeatureCollection(data []byte) (*FeatureCollection, error) {
	var stationPrice StationPriceResponse
	err := json.Unmarshal(data, &stationPrice)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%#v\n", stationPrice)

	stationData := make(map[int]StationPriceData)
	for _, st := range stationPrice.Stations {
		stationData[st.Code] = StationPriceData{
			Lat:     st.Location.Lat,
			Lon:     st.Location.Lon,
			Name:    st.Name,
			Address: st.Address,
		}
	}
	for _, pr := range stationPrice.Prices {
		sd := stationData[pr.StationCode]
		sd.Price = pr.Price
		sd.PriceUnit = pr.PriceUnit
		sd.FuelType = pr.FuelType
		sd.LastUpdated = pr.LastUpdated
		stationData[pr.StationCode] = sd
	}

	//fmt.Printf("%#v\n", stationData)

	fc := FeatureCollection{Type: "FeatureCollection"}

	feat := make([]Feature, 0)
	for _, v := range stationData {
		ft := Feature{Type: "Feature"}

		geom := Geometry{Type: "Point"}
		geom.Coordinates = make([]float64, 2)
		geom.Coordinates[0] = v.Lon
		geom.Coordinates[1] = v.Lat

		ft.Geometry = geom

		ft.Properties = make(Properties)
		ft.Properties["address"] = v.Address
		ft.Properties["price"] = v.Price
		ft.Properties["fueltype"] = v.FuelType

		feat = append(feat, ft)
	}
	fc.Features = feat

	return &fc, nil
}
