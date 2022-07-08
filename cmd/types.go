package cmd

type Location struct {
	Distance float64 `json:"distance"`
	Lat      float64 `json:"latitude"`
	Lon      float64 `json:"longitude"`
}

type Station struct {
	Code     int      `json:"code"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type StationPrice struct {
	StationCode int     `json:"stationcode"`
	Price       float64 `json:"price"`
	PriceUnit   string  `json:"priceunit"`
	FuelType    string  `json:"fueltype"`
	LastUpdated string  `json:"lastupdated"`
}

type StationPriceResponse struct {
	Stations []Station      `json:"stations"`
	Prices   []StationPrice `json:"prices"`
}

type StationPriceData struct {
	Lat         float64 `json:"latitude"`
	Lon         float64 `json:"longitude"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	Price       float64 `json:"price"`
	PriceUnit   string  `json:"priceunit"`
	FuelType    string  `json:"fueltype"`
	LastUpdated string  `json:"lastupdated"`
}

type Properties map[string]interface{}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}
