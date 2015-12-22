// Structure of Item
package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math"
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
	WordMap      []MapItem
	DictVersion  int
	Cluster      mgo.DBRef
	vector       map[int]float64
}

type MapItem struct {
	Word  string
	Freq  int
	Morph string
}

// Calculate vector
func (this *Item) calcVector() {
	if this.vector == nil {
		this.vector = make(map[int]float64)

		var tf float64
		var idf float64
		//lenght := len(this.WordMap)

		for _, value := range this.WordMap {
			tf = 0.5 + 0.5*float64(value.Freq)/float64(this.getMaxWordFreq())
			idf = float64(math.Log(float64(dict_version.Documents) / float64(dictionary[DictKey{this.Lang, value.Word}].Cnt)))

			this.vector[dictionary[DictKey{this.Lang, value.Word}].Index] = tf * idf
		}
	}
}

func (this *Item) getMaxWordFreq() int {
	var max_freq int = 0

	for _, value := range this.WordMap {
		if value.Freq > max_freq {
			max_freq = value.Freq
		}
	}

	return max_freq
}
