package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var startTime time.Time
var db tracksMongoDB
var tickerCap *int

func getPort() string {
	return ":8080"
}

func main() {
	tickerCap = flag.Int("tickerCap", 5, "an int")
	flag.Parse()

	fmt.Printf("Running paragliding with tickerCap %d\n", *tickerCap)
	startTime = time.Now()
	db.init()

	r := mux.NewRouter()
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
	r.HandleFunc("/paragliding/webhook/new_track/{webhookid}", webhookNewtrackIDGet).Methods("GET")
	r.HandleFunc("/admin/api/tracks_count", adminTrackcountGet).Methods("GET")
	r.HandleFunc("/admin/api/tracks", adminTracksDelete).Methods("DELETE")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1" + getPort(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}
