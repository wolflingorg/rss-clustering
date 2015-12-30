package main

import (
	"flag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strings"
	"time"
)

var (
	CONFIG_PATH  string                // PATH to ini file
	config       = new(Config)         // Config struct
	db           *mgo.Database         // Data Base
	dict_version *DictionaryVersion    // Lasted dictionary version
	dictionary   map[DictKey]DictValue // Local dictionary map
	LogError     *log.Logger           // Error logger
	LogInfo      *log.Logger           // Info logger
)

func main() {
	// get flags
	flag.StringVar(&CONFIG_PATH, "c", "", "PATH to ini file")
	flag.Parse()

	// config
	LoadConfig(config, CONFIG_PATH)

	// log file
	f, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %s\n", err)
	}

	// loggers
	LogInfo = log.New(f,
		"INFO: ",
		log.Ldate|log.Ltime)
	LogError = log.New(f,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// connect to db
	session, err := mgo.Dial(strings.Join(config.Db.Host, ","))
	if err != nil {
		LogError.Fatalf("Couldnt connect to mongodb server %s", err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db = session.DB("rss")
	n := db.C("news")
	c := db.C("clusters")

	// current dict version
	dict_version = GetLastDictionaryVersion()

	// dictionary map
	dictionary = make(map[DictKey]DictValue)
	UpdateDictionaryMap()

	for {
		// update dictionary
		tmp := GetLastDictionaryVersion()
		if tmp.Id > dict_version.Id {
			dict_version = tmp
			UpdateDictionaryMap()
		}

		// get tasks
		items := GetTasks()
		if len(items) > 0 {
			// get ckusters with items wordchecksums
			clusters := GetClusters(GetWordchecksum(items))

			for i, item := range items {
				var cur_cluster *Cluster

				// get item vector
				item_vector := calcVector(item.Lang, item.WordMap)
				// try to get same clusters
				same_clusters := GetSameClusters(item.WordChecksum, clusters)
				if len(same_clusters) > 0 {
					// try to find cluster
					cur_cluster = GetCurCluster(item_vector, same_clusters)
				}

				if cur_cluster != nil {
					UpdateClusterByItem(cur_cluster, item)
				} else {
					cur_cluster = NewClusterByItem(item)

					clusters = append(clusters, *cur_cluster)
					items[i].Cluster = mgo.DBRef{
						Collection: "clusters",
						Id:         cur_cluster.Id,
					}
				}
			}

			// update news
			for _, item := range items {
				n.Update(bson.M{"_id": item.Id}, bson.M{"$set": bson.M{
					"cluster":     item.Cluster,
					"clusterdate": time.Now(),
					"status":      3,
				}})
			}

			// update clusters
			for _, cluster := range clusters {
				var doc interface{}

				change := mgo.Change{
					Update: bson.M{
						"$set": bson.M{
							"_id":          cluster.Id,
							"lang":         cluster.Lang,
							"items":        cluster.Items,
							"date":         cluster.Date,
							"wordmap":      cluster.WordMap,
							"wordchecksum": cluster.WordChecksum,
							"main.title":   cluster.Main.Title,
							"main.content": cluster.Main.Content,
							"main.image":   cluster.Main.Image,
							"main.news":    cluster.Main.News,
							"news":         cluster.News,
						},
					},
					Upsert: true,
				}
				c.Find(bson.M{"_id": cluster.Id}).Apply(change, &doc)
			}

			LogInfo.Printf("%d tasks updated", len(items))
		} else {
			LogInfo.Println("No tasks")
			time.Sleep(time.Second * 5)
		}
	}
}

// get news for clustering
func GetTasks() []Item {
	n := db.C("news")
	var items []Item

	err := n.Find(bson.M{
		"dictversion": bson.M{"$lte": dict_version.Id},
		"status":      2,
	}).Sort("date").Limit(config.Handler.Tasks).All(&items)
	if err != nil {
		LogError.Fatalf("Couldnt get mongodb result %s\n", err)
	}

	return items
}
