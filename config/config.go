package config

type NswFuelApiConfig struct {
	ApiKey    string
	ApiSecret string
}

type MapifyConfig struct {
	ApiKey string
}

type Configuration struct {
	NswFuelApi NswFuelApiConfig
	Mapify     MapifyConfig
}
