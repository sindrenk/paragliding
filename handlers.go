package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/gorilla/mux"

	"github.com/marni/goigc"
)

type metainfo struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/paragliding/api/", http.StatusSeeOther)
}

func infoGetHandler(w http.ResponseWriter, r *http.Request) {
	info := metainfo{
		Uptime:  time.Since(startTime).String(),
		Info:    "Service for IGC tracks",
		Version: "v1"}
	js, err := json.Marshal(info)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func trackPostHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	url := make(map[string]string)
	err := decoder.Decode(&url)

	// couldn't parse POST data
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trackIGC, err := igc.ParseLocation(url["url"])

	// couldn't get track from url, probably a bad URL in POST request
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var dist float64

	for i := 1; i < len(trackIGC.Points); i++ {
		dist += trackIGC.Points[i-1].Distance(trackIGC.Points[i])
	}

	id, err := db.add(track{
		Hdate:       trackIGC.Date,
		Pilot:       trackIGC.Pilot,
		Glider:      trackIGC.GliderType,
		GliderID:    trackIGC.GliderID,
		TrackLength: dist,
		TrackSrcURL: url["url"],
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapid := make(map[string]string)

	mapid["id"] = id.Hex()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapid)
}

func trackGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(db.getAllIds())
}

func trackIDGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	track, ok := db.get(bson.ObjectIdHex(vars["trackid"]))

	if !ok {
		http.NotFound(w, r)
		return
	}

	js, err := json.Marshal(track)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func trackIDFieldGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	track, ok := db.get(bson.ObjectIdHex(vars["trackid"]))

	if !ok {
		http.NotFound(w, r)
		return
	}

	fieldVal, ok := track.getField(vars["field"])

	if !ok {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, fieldVal)

}

func tickerLatestGetHandler(w http.ResponseWriter, r *http.Request) {

}

func tickerGetHandler(w http.ResponseWriter, r *http.Request) {

}

func tickerTimestampGetHandler(w http.ResponseWriter, r *http.Request) {

}

func webhookNewtrackPostHandler(w http.ResponseWriter, r *http.Request) {

}

func webhookNewtrackIDGet(w http.ResponseWriter, r *http.Request) {

}

func adminTrackcountGet(w http.ResponseWriter, r *http.Request) {

}

func adminTracksDelete(w http.ResponseWriter, r *http.Request) {

}
