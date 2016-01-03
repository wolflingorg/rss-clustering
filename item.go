// Structure of Item
package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Item struct {
	Id           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Title        string
	Summary      string
	Content      string
	Link         string
	Image        *Image
	Date         time.Time
	Lang         string
	WordChecksum []string
	WordMap      []WordMapItem
	DictVersion  int
	Cluster      mgo.DBRef
	ClusterDate  time.Time
	Status       uint
	Category     string
	Country      string
}

type WordMapItem struct {
	Word  string
	Freq  int
	Morph string
}
