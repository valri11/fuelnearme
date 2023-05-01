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
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valri11/fuelnearme/config"
	"github.com/valri11/fuelnearme/geocode"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Fuel watch web server",
	Long:  `Fuel watch web server`,
	Run:   doWebServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().BoolP("dev-mode", "", false, "development mode (http on loclahost)")
	serverCmd.Flags().String("tls-cert", "", "TLS certificate file")
	serverCmd.Flags().String("tls-cert-key", "", "TLS certificate key file")
	serverCmd.Flags().Int("port", 8010, "service port to listen")
}

type fuelHandler struct {
	cfg config.Configuration
	tm  *nswApiTokenManager

	reqCounter int
}

func doWebServer(cmd *cobra.Command, args []string) {
	devMode, err := cmd.Flags().GetBool("dev-mode")
	if err != nil {
		log.Fatalf("ERR: %v", err)
		return
	}

	servicePort, err := cmd.Flags().GetInt("port")
	if err != nil {
		log.Fatalf("ERR: %v", err)
		return
	}

	tlsCertFile, err := cmd.Flags().GetString("tls-cert")
	if err != nil {
		log.Fatalf("ERR: %v", err)
		return
	}

	tlsCertKeyFile, err := cmd.Flags().GetString("tls-cert-key")
	if err != nil {
		log.Fatalf("ERR: %v", err)
		return
	}

	if !devMode {
		if tlsCertFile == "" || tlsCertKeyFile == "" {
			fmt.Println("must provide TLS key and certificate")
			return
		}
	}

	t, err := newFuelHandler()
	if err != nil {
		log.Fatalf("ERR: %v", err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/fuelnearme.geojson", t.fuelNearbyHandler)

	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "content-type", "username", "password", "Referer"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// start server listen with error handling
	mux := handlers.CORS(originsOk, headersOk, methodsOk)(r)
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", servicePort),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if devMode {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS(tlsCertFile, tlsCertKeyFile)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func newFuelHandler() (*fuelHandler, error) {
	var cfg config.Configuration
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	tm, err := NewNswApiTokenManager(cfg.NswFuelApi.ApiKey, cfg.NswFuelApi.ApiSecret)
	if err != nil {
		return nil, err
	}

	t := fuelHandler{
		cfg: cfg,
		tm:  tm,
	}

	return &t, nil
}

func (h *fuelHandler) fuelNearbyHandler(w http.ResponseWriter, r *http.Request) {

	apiToken, err := h.tm.GetOrRenewApiToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lat := -33.68769534564702
	lon := 151.1055864
	radius := 5.0
	fuelType := "E10"

	if ftParam, ok := r.URL.Query()["fueltype"]; ok {
		fuelType = ftParam[0]
	}

	if latParam, ok := r.URL.Query()["lat"]; ok {
		lat, err = strconv.ParseFloat(latParam[0], 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if lonParam, ok := r.URL.Query()["lon"]; ok {
		lon, err = strconv.ParseFloat(lonParam[0], 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if radiusParam, ok := r.URL.Query()["radius"]; ok {
		radius, err = strconv.ParseFloat(radiusParam[0], 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addr, err := geocode.AddressFromLatLon(lat, lon, h.cfg.Mapify.ApiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fuelReq := NewFuelNearmeReq()

	fuelReq.FuelType = fuelType
	fuelReq.Lat = lat
	fuelReq.Lon = lon
	fuelReq.Radius = radius
	fuelReq.NamedLocation = addr.PostCode

	data, err := json.Marshal(fuelReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(data)

	reqUrl := "https://api.onegov.nsw.gov.au/FuelPriceCheck/v2/fuel/prices/nearby"

	client := http.Client{}

	req, err := http.NewRequest("POST", reqUrl, reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqTimestamp := time.Now().Format("02/01/2006 15:04:05")

	req.Header = http.Header{
		"Content-Type":     {"application/json"},
		"Authorization":    {fmt.Sprintf("Bearer %s", apiToken)},
		"apikey":           {h.cfg.NswFuelApi.ApiKey},
		"transactionid":    {fmt.Sprintf("tr_%5d", reqCounter)},
		"requesttimestamp": {reqTimestamp},
	}
	h.reqCounter++

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fc, err := fuelNearmeDataToFeatureCollection(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := json.Marshal(fc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
