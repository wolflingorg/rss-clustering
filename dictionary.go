package main

import (
	"sync"
)

var dict_mutex = &sync.Mutex{}

type Dictionary struct {
	Lang string
	Word string
	Cnt  int
}

type DictKey struct {
	Lang, Word string
}

type DictValue struct {
	Index, Cnt int
}

func UpdateDictionaryMap() {
	d := db.C("dictionary")

	var words []Dictionary

	err := d.Find(nil).Sort("_id").All(&words)
	if err != nil {
		LogError.Fatalf("Couldnt load dictionary. Something wrong with mongodb: %s\n", err)
	}

	dict_mutex.Lock()
	for i, item := range words {
		dictionary = make(map[DictKey]DictValue)
		dictionary[DictKey{item.Lang, item.Word}] = DictValue{i, item.Cnt}
	}
	dict_mutex.Unlock()

	words = nil

	LogInfo.Printf("Dictionary version updated\n")
}
