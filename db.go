package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type tracksMongoDB struct {
	DatabaseURL          string `yaml:"database_url"`
	DatabaseName         string `yaml:"database_name"`
	TracksCollectionName string `yaml:"tracks_collection_name"`
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

	index := mgo.Index{
		Key:        []string{"track_src_url"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func (db *tracksMongoDB) add(t track) (bson.ObjectId, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Insert(t)
	if err != nil {
		fmt.Printf("Error in add(): %v", err.Error())
		return "error", err
	}

	track := track{}
	err = session.DB(db.DatabaseName).C(db.TracksCollectionName).Find(bson.M{"track_src_url": t.TrackSrcURL}).One(&track)
	if err != nil {
		fmt.Printf("Error in add(): %v", err.Error())
		return "error", err
	}

	return track.ID, nil
}

func (db *tracksMongoDB) count() int {
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

func (db *tracksMongoDB) get(id bson.ObjectId) (track, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := track{}
	found := false

	for _, validID := range db.getAllIds() {
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

func (db *tracksMongoDB) getAllIds() []bson.ObjectId {
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

func (db *tracksMongoDB) getField(field string, id bson.ObjectId) (string, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := track{}

	found := false

	for _, validID := range db.getAllIds() {
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
