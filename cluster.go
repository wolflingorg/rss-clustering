// Structure of Cluster
package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var cluster_mutex = &sync.Mutex{}

type Cluster struct {
	Id           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Lang         string
	Date         time.Time
	Items        int
	WordMap      []WordMapItem
	WordChecksum []string
	Main         ClusterMainNews
	News         []mgo.DBRef
	vector       map[int]float64
}

type ClusterMainNews struct {
	Title   string
	Content string
	Image   *Image
	News    mgo.DBRef
}

func (this *Cluster) GetVector() map[int]float64 {
	if this.vector == nil {
		cluster_mutex.Lock()
		this.vector = calcVector(this.Lang, this.WordMap)
		cluster_mutex.Unlock()
	}

	return this.vector
}

func (this *Cluster) ResetVector() {
	cluster_mutex.Lock()
	this.vector = nil
	cluster_mutex.Unlock()
}
