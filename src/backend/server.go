package main

import (
	"github.com/joho/godotenv"
	"os"
	"time"
	"log"
	"net/http"
	"net/url"
	"github.com/rs/cors"
	"encoding/json"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2"
)

type Payload struct {
	EventType string
	ClickTarget string
	ClosestLinkHref string
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
	GPU string
}

func (ich *InfluxClientHandler) Monitor(w http.ResponseWriter, r *http.Request) {
	var p Payload
	err := json.NewDecoder(r.Body).Decode(&p)
	if err == nil {
		CurrentPage, error := url.Parse(p.Href)
		var QSource = ""
		var QMedium = ""
		var QCampaign = ""
		var QTerm = ""
		var QContent = ""
		if error == nil {			
			QSource = CurrentPage.Query().Get("utm_source")
			QMedium = CurrentPage.Query().Get("utm_medium")
			QCampaign = CurrentPage.Query().Get("utm_campaign")
			QTerm = CurrentPage.Query().Get("utm_term")
			QContent = CurrentPage.Query().Get("utm_content")
		} 

		tags := map[string]string{
			"website_id": p.NId,	
			"visitor_ip": r.RemoteAddr,
			"event_type": p.EventType,
		}
		fields := map[string]interface{}{
			"click_target": p.ClickTarget,
			"closest_link_href": p.ClosestLinkHref,
			"user_agent": p.UserAgent,
			"current_page": p.Href,
			"referrer": p.Referrer,
			"cpu_cores": p.HardwareConcurrency,
			"device_memory": p.DeviceMemory,
			"max_touch_points": p.MaxTouchPoints,
			"language": p.Language,
			"cookie_support": p.CookieEnabled,
			"pdf_support": p.PDFViewerEnabled,
			"online": p.OnLine,
			"gpu": p.GPU,
			"q_source": QSource,
			"q_medium": QMedium,
			"q_campaign": QCampaign,
			"q_term": QTerm,
			"q_content": QContent,
		}
	
		influx_p := influxdb2.NewPoint("event",
			tags,
			fields,
			time.Now())
		ich.WriteAPI.WritePoint(influx_p)
		ich.WriteAPI.Flush() // should probably be removed after testing, might have performance impact
	}
	w.WriteHeader(http.StatusOK)
}

func (ich *InfluxClientHandler) MonitorPixel(w http.ResponseWriter, r *http.Request) {
	QNid := r.URL.Query().Get("nid")
	if QNid != "" {
		tags := map[string]string{
			"website_id": QNid,	
			"visitor_ip": r.RemoteAddr,
			"event_type": "page_pixel",
		}
		fields := map[string]interface{}{
			"ping": true,
		}

		influx_p := influxdb2.NewPoint("event",
			tags,
			fields,
			time.Now())
		ich.WriteAPI.WritePoint(influx_p)
		ich.WriteAPI.Flush() // should probably be removed after testing, might have performance impact
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
	log.Println("nuunamnir website analytics tracking server started ...")
	godotenv.Load("../../data/input/credentials.env")
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
	mux.HandleFunc("/monitor.gif", myInfluxClientHandler.MonitorPixel)
	mux.HandleFunc("/favicon.ico", FavIcon)
	handler := c.Handler(mux)

	tlsCert = os.Getenv("TLS_CERT")
	tlsKey = os.Getenv("TLS_KEY")
	http.ListenAndServeTLS(":3106", tlsCert, tlsKey, handler)

	defer client.Close()
}
