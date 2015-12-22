// Structure of Cluster
package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Cluster struct {
	Id    bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Lang  string
	Date  time.Time
	Items int
	Main  ClusterMainNews
	News  []mgo.DBRef
}

type ClusterMainNews struct {
	Title   string
	Content string
	Image   *Image
	News    mgo.DBRef
}
