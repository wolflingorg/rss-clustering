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
	Title    string
	Content  string
	Image    *Image
	Category string
	Country  string
	News     mgo.DBRef
}

type ClasterAngle struct {
	angle   float64
	cluster *Cluster
}

// try to get vector for cluster
// if vector doesnt exists - calc it
func (this *Cluster) GetVector() map[int]float64 {
	if this.vector == nil {
		cluster_mutex.Lock()
		this.vector = calcVector(this.Lang, this.WordMap)
		cluster_mutex.Unlock()
	}

	return this.vector
}

// reset vector after update wordmap
func (this *Cluster) ResetVector() {
	cluster_mutex.Lock()
	this.vector = nil
	cluster_mutex.Unlock()
}

// try to find cluster for news
// - calc vector for cluster
// - calc angle between news vector and cluster vector
// - compares the obtained angle with the threshold value
func GetCurCluster(item_vector map[int]float64, same_clusters []*Cluster) *Cluster {
	var max_angle float64
	var cur_cluster *Cluster

	sync_chan := make(chan ClasterAngle)

	for i, _ := range same_clusters {
		go func(item_vector map[int]float64, cluster *Cluster) {
			cluster_vector := cluster.GetVector()
			sync_chan <- ClasterAngle{
				angle:   calcAngle(item_vector, cluster_vector),
				cluster: cluster,
			}
		}(item_vector, same_clusters[i])
	}

	for i := 0; i < len(same_clusters); i++ {
		select {
		case result := <-sync_chan:
			if result.angle >= config.Clustering.Porog && result.angle > max_angle {
				max_angle = result.angle
				cur_cluster = result.cluster
			}
		}
	}

	return cur_cluster
}

// get clusters from DB with same wordchecksum from all news
func GetClusters(wordchecksum []string) []Cluster {
	c := db.C("clusters")
	var clusters []Cluster

	err := c.Find(bson.M{
		"wordchecksum": bson.M{"$in": wordchecksum},
	}).All(&clusters)
	if err != nil {
		LogError.Fatalf("Couldnt get mongodb result %s\n", err)
	}

	return clusters
}

// try to find clusters with same wordchecksum from curent news
func GetSameClusters(wordchecksum []string, clusters []Cluster) []*Cluster {
	var same_clusters []*Cluster

	for i, cluster := range clusters {
		for _, value := range wordchecksum {
			if stringInSlice(value, cluster.WordChecksum) {
				same_clusters = append(same_clusters, &clusters[i])
				break
			}
		}
	}

	return same_clusters
}

// update cluster
func UpdateClusterByItem(cur_cluster *Cluster, item Item) {
	wordmap, wordchecksum := GetTopWords(item.Lang, AppendWordMap(cur_cluster.WordMap, item.WordMap))

	cur_cluster.Date = item.Date
	cur_cluster.Items = cur_cluster.Items + 1
	cur_cluster.Main = ClusterMainNews{
		Title:    item.Title,
		Content:  item.Content,
		Image:    item.Image,
		Category: item.Category,
		Country:  item.Country,
		News: mgo.DBRef{
			Collection: "news",
			Id:         item.Id,
		},
	}
	cur_cluster.News = append(cur_cluster.News, mgo.DBRef{
		Collection: "news",
		Id:         item.Id,
	})
	cur_cluster.WordMap = wordmap
	cur_cluster.WordChecksum = wordchecksum
	cur_cluster.ResetVector()
}

// create new cluster
func NewClusterByItem(item Item) *Cluster {
	cur_cluster := &Cluster{
		Id:    bson.NewObjectId(),
		Lang:  item.Lang,
		Date:  item.Date,
		Items: 1,
		Main: ClusterMainNews{
			Title:    item.Title,
			Content:  item.Content,
			Image:    item.Image,
			Category: item.Category,
			Country:  item.Country,
			News: mgo.DBRef{
				Collection: "news",
				Id:         item.Id,
			},
		},
	}
	cur_cluster.News = append(cur_cluster.News, mgo.DBRef{
		Collection: "news",
		Id:         item.Id,
	})
	cur_cluster.WordMap = item.WordMap
	cur_cluster.WordChecksum = item.WordChecksum

	return cur_cluster
}
