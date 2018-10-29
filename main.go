package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var startTime time.Time
var db tracksMongoDB
var tickerCap *int

func getPort() string {
	if os.Getenv("PORT") == "" {
		return ":8080"
	} else {
		return ":" + os.Getenv("PORT")
	}
}

func main() {
	tickerCap = flag.Int("tickerCap", 5, "an int")
	flag.Parse()

	db.init()

	fmt.Printf("Running paragliding with tickerCap %d\n", *tickerCap)
	startTime = time.Now()

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/paragliding", rootHandler)
	r.HandleFunc("/paragliding/api", infoGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/track", trackPostHandler).Methods("POST")
	r.HandleFunc("/paragliding/api/track", trackGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/track/{trackid}", trackIDGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/track/{trackid}/{field}", trackIDFieldGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/ticker/latest", tickerLatestGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/ticker", tickerGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/ticker/{timestamp}", tickerTimestampGetHandler).Methods("GET")
	r.HandleFunc("/paragliding/api/webhook/new_track", webhookNewtrackPostHandler).Methods("POST")
	r.HandleFunc("/paragliding/api/webhook/new_track/{webhookid}", webhookNewtrackIDGet).Methods("GET")
	r.HandleFunc("/paragliding/api/webhook/new_track/{webhookid}", webhookNewtrackIDDelete).Methods("DELETE")
	r.HandleFunc("/admin/api/tracks_count", adminTrackcountGet).Methods("GET")
	r.HandleFunc("/admin/api/tracks", adminTracksDelete).Methods("DELETE")

	log.Fatal(http.ListenAndServe(getPort(), r))

}
