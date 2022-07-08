package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReadPricesNearby(t *testing.T) {
	testFilePath := filepath.Join("testdata", "price_nearby.json")
	data, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("error reading source file:", err)
	}
	//fmt.Printf("%s\n", string(data))

	var stationPrice StationPriceResponse
	err = json.Unmarshal(data, &stationPrice)
	if err != nil {
		t.Fatal("error unmarshal:", err)
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

	fmt.Printf("%#v\n", stationData)

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

	//fmt.Printf("%#v\n", fc)

	out, err := json.Marshal(fc)
	if err != nil {
		t.Fatal("error marshal:", err)
	}
	fmt.Printf("%s\n", string(out))
}
