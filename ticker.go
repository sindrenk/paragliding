package main

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
)

type ticker struct {
	TLatest    int64           `json:"t_latest"`
	TStart     int64           `json:"t_start"`
	TStop      int64           `json:"t_stop"`
	Tracks     []bson.ObjectId `json:"tracks"`
	Processing int64           `json:"processing"`
}

func replyWithTicker(timestamp ...int64) (ticker, bool) {
	timeBefore := time.Now()
	if len(timestamp) > 1 {
		panic(errors.New("replyWithTicker() must have 0 or 1 arguments"))
	} else if len(timestamp) == 0 { // if function doesn't have variables set time to earliest possible
		timestamp = append(timestamp, 0)
	}

	var ticker ticker
	tracks := db.getTracksSince(*tickerCap, time.Unix(timestamp[0], 0))

	if len(tracks) <= 0 {
		return ticker, false
	}

	ticker.TLatest = db.getLatest().ID.Time().Unix()
	ticker.TStart = tracks[0].ID.Time().Unix()
	ticker.TStop = tracks[0].ID.Time().Unix()

	for _, track := range tracks {
		ticker.Tracks = append(ticker.Tracks, track.ID)
		if track.ID.Time().Unix() < ticker.TStop {
			ticker.TStop = track.ID.Time().Unix()
		}

		if track.ID.Time().Unix() > ticker.TStart {
			ticker.TStart = track.ID.Time().Unix()
		}
	}

	ticker.Processing = int64(time.Since(timeBefore).Seconds() * 1000)

	return ticker, true
}
