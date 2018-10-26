package main

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
)

type track struct {
	Hdate       time.Time     `json:"H_date" bson:"H_date"`
	Pilot       string        `json:"pilot" bson:"pilot"`
	Glider      string        `json:"glider" bson:"glider"`
	GliderID    string        `json:"glider_id" bson:"glider_id"`
	TrackLength float64       `json:"track_length" bson:"track_length"`
	TrackSrcURL string        `json:"track_src_url" bson:"track_src_url"`
	ID          bson.ObjectId `bson:"_id,omitempty"`
}

// Returns the contents of a specific field, as is named in JSON
func (t *track) getField(field string) (string, bool) {
	var f string
	ok := true
	switch field {
	case "H_date":
		f = t.Hdate.String()
	case "pilot":
		f = t.Pilot
	case "glider":
		f = t.Glider
	case "glider_id":
		f = t.GliderID
	case "track_length":
		f = fmt.Sprintf("%f", t.TrackLength)
	default:
		f = "INVALID FIELD"
		ok = false
	}

	return f, ok
}
