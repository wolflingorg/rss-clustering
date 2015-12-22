package main

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
		panic(err)
	}

	for i, item := range words {
		dictionary[DictKey{item.Lang, item.Word}] = DictValue{i, item.Cnt}
	}

	words = nil
}
