package main

import (
	"math"
)

type Vector struct {
	Id   int
	Freq float64
}

type VectorByFreq []Vector

func (a VectorByFreq) Len() int           { return len(a) }
func (a VectorByFreq) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a VectorByFreq) Less(i, j int) bool { return a[i].Freq < a[j].Freq }

// calc TF*IDF vector for news
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

// calc angle between ywo vectors
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
