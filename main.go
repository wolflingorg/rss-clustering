package main

import (
	"flag"
	//"fmt"
	"gopkg.in/mgo.v2"
	"strings"
	tm "task-manager"
	"time"
)

var (
	CONFIG_PATH  string                // PATH to ini file
	config       = new(Config)         // Config struct
	db           *mgo.Database         // Data Base
	dict_version *DictionaryVersion    // Lasted dictionary version
	dictionary   map[DictKey]DictValue // Local dictionary map
)

func main() {
	// get flags
	flag.StringVar(&CONFIG_PATH, "c", "", "PATH to ini file")
	flag.Parse()

	// config
	LoadConfig(config, CONFIG_PATH)

	// connect to db
	session, err := mgo.Dial(strings.Join(config.Db.Host, ","))
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db = session.DB("rss")

	// start task manager
	tm.StartDispatcher(config.Handler.Workers, ClusteringHandler)

	// current dict version
	dict_version = GetLastDictionaryVersion()

	// dictionary map
	dictionary = make(map[DictKey]DictValue)
	UpdateDictionaryMap()

	for {
		select {
		case <-time.After(config.Handler.Interval):
			AddTasksHandler()
		}
	}
}
