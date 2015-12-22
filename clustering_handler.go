package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math"
	tm "task-manager"
)

func ClusteringHandler(work tm.WorkRequest, worker_id int) {
	n := db.C("news")
	c := db.C("clusters")

	// check that work.Data equal Item interface
	if item, ok := work.Data.(Item); ok {
		// calc vector
		item.calcVector()

		// find similar news
		var items []Item
		err := n.Find(bson.M{
			"dictversion":  bson.M{"$exists": true, "$lte": dict_version.Id},
			"cluster":      bson.M{"$exists": true},
			"wordchecksum": bson.M{"$in": item.WordChecksum},
			"_id":          bson.M{"$ne": item.Id},
		}).All(&items)
		if err != nil {
			panic(err)
		}

		// find cluster id
		var cluster_id bson.ObjectId
		var max_angle float64

		for _, value := range items {
			value.calcVector()
			angle := getAngle(item.vector, value.vector)

			if angle >= config.Clustering.Porog {
				if angle > max_angle {
					if id, ok := value.Cluster.Id.(bson.ObjectId); ok {
						cluster_id = id
						max_angle = angle
					}
				}
			}
		}

		// if cluster id exists - add news to cluster or create atomic cluster
		if cluster_id.Valid() == true {
			c.Update(bson.M{"_id": cluster_id}, bson.M{"$set": bson.M{
				"date":         item.Date,
				"main.title":   item.Title,
				"main.content": item.Content,
				"main.image":   item.Image,
				"main.news": mgo.DBRef{
					Collection: "news",
					Id:         item.Id,
				},
			}, "$push": bson.M{
				"news": mgo.DBRef{
					Collection: "news",
					Id:         item.Id,
				},
			},
				"$inc": bson.M{"items": 1}})
		} else {
			cluster_id = bson.NewObjectId()

			cluster := Cluster{
				Id:    cluster_id,
				Lang:  item.Lang,
				Date:  item.Date,
				Items: 1,
				Main: ClusterMainNews{
					Title:   item.Title,
					Content: item.Content,
					Image:   item.Image,
					News: mgo.DBRef{
						Collection: "news",
						Id:         item.Id,
					},
				},
			}
			cluster.News = append(cluster.News, mgo.DBRef{
				Collection: "news",
				Id:         item.Id,
			})

			c.Insert(&cluster)
		}

		// update news
		n.Update(bson.M{"_id": item.Id}, bson.M{"$set": bson.M{
			"cluster": mgo.DBRef{
				Collection: "clusters",
				Id:         cluster_id,
			},
		}})

		fmt.Printf("\tWorker %d OK\n", worker_id)
	}
}

func getAngle(v1 map[int]float64, v2 map[int]float64) float64 {
	var v1v2 float64
	var modv1 float64
	var modv2 float64

	for i, value := range v1 {
		if item, ok := v2[i]; ok {
			v1v2 += value * item
		}
	}

	for _, value := range v1 {
		modv1 += value * value
	}
	modv1 = math.Sqrt(modv1)

	for _, value := range v2 {
		modv2 += value * value
	}
	modv2 = math.Sqrt(modv2)

	return v1v2 / (modv1 * modv2)
}
