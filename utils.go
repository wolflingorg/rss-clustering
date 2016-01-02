package main

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
)

// Return slice of md5 checksums of words from Item.WordMap
func GetWordchecksum(items []Item) []string {
	var wordchecksum []string

	for _, item := range items {
		wordchecksum = append(wordchecksum, item.WordChecksum...)
	}

	return wordchecksum
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

// return max word frequency from slice of WordMap
func getMaxWordFreq(wordmap []WordMapItem) int {
	var max_freq int = 0

	for _, value := range wordmap {
		if value.Freq > max_freq {
			max_freq = value.Freq
		}
	}

	return max_freq
}

// Concat two wordmaps
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

// Find WordMap by word
func findInWordMapByWord(word string, wm []WordMapItem) (*WordMapItem, bool) {
	for i, value := range wm {
		if value.Word == word {
			return &wm[i], true
		}
	}

	return &WordMapItem{}, false
}

// Return TOP WordMap and WordChecksum by TF*IDF weight
func GetTopWords(lang string, wordmap []WordMapItem) (res_wordmap []WordMapItem, res_wordchecksum []string) {
	max_words := 25

	var vectors []Vector

	vocabulary := make(map[int]string)
	for _, value := range wordmap {
		vocabulary[dictionary[DictKey{lang, value.Word}].Index] = value.Word
	}

	wordmap_by_vocalulary := make(map[int]WordMapItem)
	for _, value := range wordmap {
		wordmap_by_vocalulary[dictionary[DictKey{lang, value.Word}].Index] = value
	}

	for i, value := range calcVector(lang, wordmap) {
		vectors = append(vectors, Vector{
			Id:   i,
			Freq: value,
		})
	}
	sort.Sort(sort.Reverse(VectorByFreq(vectors)))

	if len(vectors) < max_words {
		max_words = len(vectors)
	}

	for i := 0; i < max_words; i++ {
		hasher := md5.New()
		hasher.Write([]byte(vocabulary[vectors[i].Id]))
		res_wordchecksum = append(res_wordchecksum, hex.EncodeToString(hasher.Sum(nil)))
	}

	for i := 0; i < max_words; i++ {
		res_wordmap = append(res_wordmap, wordmap_by_vocalulary[vectors[i].Id])
	}

	return
}
