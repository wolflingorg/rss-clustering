package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"math"
	"os"
	"sort"
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
			clusters := GetClusters(GetWordchecksum(items))

			for i, item := range items {
				item_vector := calcVector(item.Lang, item.WordMap)
				same_clusters := GetSameClusters(item.WordChecksum, clusters)
				var cur_cluster *Cluster

				if len(same_clusters) > 0 {
					cur_cluster = GetCurCluster(item_vector, same_clusters)
				}

				if cur_cluster != nil {
					wordmap := AppendWordMap(cur_cluster.WordMap, item.WordMap)

					cur_cluster.Date = item.Date
					cur_cluster.Items = cur_cluster.Items + 1
					cur_cluster.Main = ClusterMainNews{
						Title:   item.Title,
						Content: item.Content,
						Image:   item.Image,
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
					cur_cluster.WordChecksum = GetTopWordChecksum(item.Lang, wordmap)
				} else {
					cur_cluster = &Cluster{
						Id:    bson.NewObjectId(),
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
					cur_cluster.News = append(cur_cluster.News, mgo.DBRef{
						Collection: "news",
						Id:         item.Id,
					})
					cur_cluster.WordMap = item.WordMap
					cur_cluster.WordChecksum = item.WordChecksum

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

type ClasterAngle struct {
	angle   float64
	cluster *Cluster
}

func GetCurCluster(item_vector map[int]float64, same_clusters []*Cluster) *Cluster {
	var max_angle float64
	var cur_cluster *Cluster

	sync_chan := make(chan ClasterAngle)

	for i, _ := range same_clusters {
		go func(item_vector map[int]float64, cluster *Cluster) {
			cluster_vector := calcVector(cluster.Lang, cluster.WordMap)
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

func GetTasks() []Item {
	n := db.C("news")
	var items []Item

	err := n.Find(bson.M{
		"dictversion": bson.M{"$lte": dict_version.Id},
		"status":      2,
		//"wordscount":  bson.M{"$gte": 50},
	}).Sort("date").Limit(config.Handler.Tasks).All(&items)
	if err != nil {
		LogError.Fatalf("Couldnt get mongodb result %s\n", err)
	}

	return items
}

func GetWordchecksum(items []Item) []string {
	var wordchecksum []string

	for _, item := range items {
		wordchecksum = append(wordchecksum, item.WordChecksum...)
	}

	return wordchecksum
}

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

// Check if string exists in slice
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func calcAngle(v1 map[int]float64, v2 map[int]float64) float64 {
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

func getMaxWordFreq(wordmap []WordMapItem) int {
	var max_freq int = 0

	for _, value := range wordmap {
		if value.Freq > max_freq {
			max_freq = value.Freq
		}
	}

	return max_freq
}

func calcVector(lang string, wordmap []WordMapItem) map[int]float64 {
	vector := make(map[int]float64)
	var tf float64
	var idf float64

	for _, value := range wordmap {
		tf = 0.5 + 0.5*float64(value.Freq)/float64(getMaxWordFreq(wordmap))
		idf = float64(math.Log(float64(dict_version.Documents) / float64(dictionary[DictKey{lang, value.Word}].Cnt)))

		vector[dictionary[DictKey{lang, value.Word}].Index] = tf * idf
	}

	return vector
}

func AppendWordMap(wm1 []WordMapItem, wm2 []WordMapItem) []WordMapItem {
	for _, value := range wm2 {
		if item, ok := findInWordMapByWord(value.Word, wm1); ok {
			item.Freq = item.Freq + value.Freq
		} else {
			wm1 = append(wm1, value)
		}
	}

	return wm1
}

func findInWordMapByWord(word string, wm []WordMapItem) (*WordMapItem, bool) {
	for i, value := range wm {
		if value.Word == word {
			return &wm[i], true
		}
	}

	return &WordMapItem{}, false
}

type Vector struct {
	Id   int
	Freq float64
}

type VectorByFreq []Vector

func (a VectorByFreq) Len() int           { return len(a) }
func (a VectorByFreq) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a VectorByFreq) Less(i, j int) bool { return a[i].Freq < a[j].Freq }

func GetTopWordChecksum(lang string, wordmap []WordMapItem) []string {
	max_checksums := 50
	var word_checksum []string
	var vectors []Vector

	vocabulary := make(map[int]string)
	for _, value := range wordmap {
		vocabulary[dictionary[DictKey{lang, value.Word}].Index] = value.Word
	}

	for i, value := range calcVector(lang, wordmap) {
		vectors = append(vectors, Vector{
			Id:   i,
			Freq: value,
		})
	}
	sort.Sort(sort.Reverse(VectorByFreq(vectors)))

	if len(vectors) < max_checksums {
		max_checksums = len(vectors)
	}

	for i := 0; i < max_checksums; i++ {
		hasher := md5.New()
		hasher.Write([]byte(vocabulary[vectors[i].Id]))
		word_checksum = append(word_checksum, hex.EncodeToString(hasher.Sum(nil)))
	}

	return word_checksum
}
