package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type tracksMongoDB struct {
	DatabaseURL            string `yaml:"database_url"`
	DatabaseName           string `yaml:"database_name"`
	TracksCollectionName   string `yaml:"tracks_collection_name"`
	WebhooksCollectionName string `yaml:"webhooks_collection_name"`
}

func (db *tracksMongoDB) init() {
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Couldn't find config.yaml")
		panic(err)
	}

	err = yaml.Unmarshal(configFile, db)
	if err != nil {
		fmt.Printf("Couldn't read from config.yaml, check format")
		panic(err)
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	trackIndex := mgo.Index{
		Key:        []string{"track_src_url"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).EnsureIndex(trackIndex)
	if err != nil {
		panic(err)
	}

	webhookIndex := mgo.Index{
		Key:        []string{"webhook_URL"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.WebhooksCollectionName).EnsureIndex(webhookIndex)
	if err != nil {
		panic(err)
	}
}

func (db *tracksMongoDB) addTrack(t track) (bson.ObjectId, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Insert(t)
	if err != nil {
		fmt.Printf("Error in addTrack(): %v", err.Error())
		return "error", err
	}

	track := track{}
	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{"track_src_url": t.TrackSrcURL}).One(&track)
	if err != nil {
		fmt.Printf("Error in addTrack(): %v", err.Error())
		return "error", err
	}

	invokeWebhooks()

	return track.ID, nil
}

func (db *tracksMongoDB) countTracks() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.TracksCollectionName).Count()
	if err != nil {
		fmt.Printf("Error in count(): %v", err.Error())
		return -1
	}

	return count
}

func (db *tracksMongoDB) getTrack(id bson.ObjectId) (track, bool) {
	if db.countTracks() <= 0 {
		fmt.Print("No entries in database")
		return track{}, false
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := track{}
	found := false

	for _, validID := range db.getTrackAllIds() {
		if validID == id {
			found = true
			break
		}
	}

	if found {
		session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{"_id": id}).One(&track)
	}

	return track, found
}

func (db *tracksMongoDB) getTrackAllIds() []bson.ObjectId {
	if db.countTracks() <= 0 {
		fmt.Print("No entries in database")
		return []bson.ObjectId{}
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	var tracks []track

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{}).All(&tracks)
	if err != nil {
		return make([]bson.ObjectId, 0, 0)
	}

	var IDList []bson.ObjectId
	for _, track := range tracks {
		IDList = append(IDList, track.ID)
	}

	return IDList
}

func (db *tracksMongoDB) getTrackField(field string, id bson.ObjectId) (string, bool) {
	if db.countTracks() <= 0 {
		fmt.Print("No entries in database")
		return "NO ENTRIES", false
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := track{}

	found := false

	for _, validID := range db.getTrackAllIds() {
		if validID == id {
			found = true
			break
		}
	}

	if !found {
		return "NOT FOUND", false
	}

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{"_id": id}).One(&track)

	// Already checked if id exists, i.e. if err != nil there is something wrong with the db
	if err != nil {
		panic(err)
	}

	return track.getField(field)
}

func (db *tracksMongoDB) getTracksSince(count int, since time.Time) []track {
	if db.countTracks() <= 0 {
		fmt.Print("No entries in database")
		return []track{}
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	tracksIter := session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{}).Iter()
	defer tracksIter.Close()

	var tracks []track
	var track track
	i := 0
	for tracksIter.Next(&track) && i < count {
		if track.ID.Time().After(since) {
			tracks = append(tracks, track)
			i++
		}
	}

	return tracks
}

func (db *tracksMongoDB) getLatestTrack() track {
	if db.countTracks() <= 0 {
		fmt.Print("No entries in database")
		return track{}
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var track track
	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(nil).Skip(db.countTracks() - 1).One(&track)
	if err != nil {
		fmt.Printf("Error in getLatest(): %v", err)
		return track
	}

	return track
}

func (db *tracksMongoDB) deleteAllTracks() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.DB(db.DatabaseName).C(db.TracksCollectionName).RemoveAll(bson.M{})
}

func (db *tracksMongoDB) addWebhook(wh webhook) (bson.ObjectId, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	wh.LatestInvokedTrack = db.countTracks()

	err = session.DB(db.DatabaseName).C(db.WebhooksCollectionName).Insert(wh)
	if err != nil {
		fmt.Printf("Error in addWebhook(): %v", err.Error())
		return "error", err
	}

	webhook := webhook{}
	err = session.DB(db.DatabaseName).C(db.WebhooksCollectionName).Find(bson.M{"webhook_URL": wh.URL}).One(&webhook)
	if err != nil {
		fmt.Printf("Error in addWebhook(): %v", err.Error())
		return "error", err
	}

	return webhook.ID, nil
}

func (db *tracksMongoDB) getWebhooksToInvoke() []webhookResponse {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	webhooksIter := session.DB(db.DatabaseName).C(db.WebhooksCollectionName).Find(bson.M{}).Iter()
	defer webhooksIter.Close()

	var webhookResponses []webhookResponse
	var webhook webhook

	allTrackIDs := db.getTrackAllIds()
	for webhooksIter.Next(&webhook) {
		timeBefore := time.Now()
		if webhook.LatestInvokedTrack+webhook.MinTriggerValue >= len(allTrackIDs) {
			var trackIDs []bson.ObjectId
			latestTrack, ok := db.getTrack(allTrackIDs[len(allTrackIDs)-1])
			if !ok {
				fmt.Println("No tracks in database")
				break
			}

			for i := webhook.LatestInvokedTrack; i < webhook.LatestInvokedTrack+webhook.MinTriggerValue; i++ {
				trackIDs = append(trackIDs, allTrackIDs[i])
			}
			webhookResponses = append(webhookResponses, webhookResponse{
				TLatest:    int64(latestTrack.ID.Time().Unix() * 1000),
				Tracks:     trackIDs,
				Processing: int64(time.Since(timeBefore).Seconds() * 1000),
				URL:        webhook.URL,
			})
		}
	}

	return webhookResponses
}

func (db *tracksMongoDB) getWebhook(id bson.ObjectId) (webhook, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	webhook := webhook{}

	err = session.DB(db.DatabaseName).C(db.WebhooksCollectionName).Find(bson.M{"_id": id}).One(&webhook)
	if err != nil {
		return webhook, false
	}

	return webhook, true
}

func (db *tracksMongoDB) deleteWebhook(id bson.ObjectId) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.WebhooksCollectionName).Remove(bson.M{"_id": id})

	return err == nil
}
