package main

import (
	"time"
)

type DictionaryVersion struct {
	Id        int `json:"id" bson:"_id,omitempty"`
	Date      time.Time
	Documents int
}

// Return last dictionary version
func GetLastDictionaryVersion() *DictionaryVersion {
	item := new(DictionaryVersion)
	dv := db.C("dictionary_versions")

	dv.Find(nil).Sort("-_id").Limit(1).One(&item)
	return item
}
