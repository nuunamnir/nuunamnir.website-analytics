package main

import (
	"github.com/joho/godotenv"
	"os"
	"time"
	"log"
	"net/http"
	"github.com/rs/cors"
	"encoding/json"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2"
)

type Payload struct {
	UserAgent string
	HardwareConcurrency int
	DeviceMemory float32
	MaxTouchPoints int
	Language string
	Href string
	Referrer string
	NId string
	CookieEnabled bool
	PDFViewerEnabled bool
	OnLine bool
}

func (ich *InfluxClientHandler) Monitor(w http.ResponseWriter, r *http.Request) {
	log.Println("Monitor:")
	log.Println(r.RemoteAddr)
	var p Payload
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Fatal("Cannot read payload.")
	}
	log.Println(p.UserAgent)
	log.Println(p.HardwareConcurrency)
	log.Println(p.DeviceMemory)
	log.Println(p.MaxTouchPoints)
	log.Println(p.Language)
	log.Println(p.Href)
	log.Println(p.Referrer)
	log.Println(p.NId)
	log.Println(p.CookieEnabled)
	log.Println(p.PDFViewerEnabled)
	log.Println(p.OnLine)
	w.WriteHeader(http.StatusOK)
	
	tags := map[string]string{
		"website_id": p.NId,	
		"visitor_ip": r.RemoteAddr,
	}
	fields := map[string]interface{}{
		"current_page": p.Href,
		"cpu_cores": p.HardwareConcurrency,
	}

	influx_p := influxdb2.NewPoint("event",
        tags,
        fields,
        time.Now())
    ich.WriteAPI.WritePoint(influx_p)
	ich.WriteAPI.Flush() // should probably be removed after testing, might have performance impact
}

func MonitorPixel(w http.ResponseWriter, r *http.Request) {
	log.Println("MonitorPixel:")
	log.Println(r.RemoteAddr)
	keys, ok := r.URL.Query()["nid"]
	if !ok {
		log.Fatal("Cannot read URL parameters.")
	} else {
		log.Println(keys[0])
	}
	w.WriteHeader(http.StatusNoContent)
}

func Status(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func FavIcon(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type InfluxClientHandler struct {
	WriteAPI api.WriteAPI
}


func main() {
	godotenv.Load("../data/input/credentials.env")
	dbToken := os.Getenv("INFLUXDB_TOKEN")
	dbURL := os.Getenv("INFLUXDB_URL")
	dbOrganization := os.Getenv("INFLUXDB_ORGANIZATION")
	dbBucket := os.Getenv("INFLUXDB_BUCKET")
	client := influxdb2.NewClient(dbURL, dbToken)
	writeAPI := client.WriteAPI(dbOrganization, dbBucket)
	myInfluxClientHandler := &InfluxClientHandler{WriteAPI: writeAPI}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		Debug: false,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/status", Status)
	mux.HandleFunc("/monitor", myInfluxClientHandler.Monitor)
	mux.HandleFunc("/monitor.gif", MonitorPixel)
	mux.HandleFunc("/favicon.ico", FavIcon)
	handler := c.Handler(mux)
	http.ListenAndServe(":3106", handler)

	defer client.Close()
}