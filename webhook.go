package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/globalsign/mgo/bson"
)

type webhook struct {
	URL                string        `json:"webhook_URL" bson:"webhook_URL"`
	MinTriggerValue    int           `json:"min_trigger_value" bson:"min_trigger_value"`
	ID                 bson.ObjectId `bson:"_id,omitempty"`
	LatestInvokedTrack int           `bson:"latest_invoked_track"`
}

type webhookResponse struct {
	TLatest    int64           `json:"t_latest"`
	Tracks     []bson.ObjectId `json:"tracks"`
	Processing int64           `json:"processing"`
	URL        string
}

func invokeWebhooks() {
	for _, wr := range db.getWebhooksToInvoke() {
		var tracksString string
		for _, track := range wr.Tracks {
			tracksString = tracksString + track.Hex() + ", "
		}
		jsonStr := "{\"text\": \"Latest timestamp: " + strconv.Itoa(int(wr.TLatest)) + ", " + strconv.Itoa(len(wr.Tracks)) + " new tracks are: " + tracksString + " (processing: " + strconv.Itoa(int(wr.Processing)) + "ms)\"}"
		fmt.Print(jsonStr)

		req, err := http.NewRequest("POST", wr.URL, bytes.NewBuffer([]byte(jsonStr)))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
	}
}
