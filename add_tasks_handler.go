package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	tm "task-manager"
)

func AddTasksHandler() {
	c := db.C("news")
	var items []Item

	// try to find feeds to update
	limit := config.Handler.Tasks - tm.GetTasksCount()
	if limit <= 0 {
		// TODO delete this
		fmt.Printf("\nTasks didnt add. %d active tasks count\n", tm.GetTasksCount())
		return
	}

	err := c.Find(bson.M{
		"dictversion": bson.M{"$exists": true, "$lte": dict_version.Id},
		"cluster":     bson.M{"$exists": false},
		"_id":         bson.M{"$nin": tm.GetTasksIds()},
	}).Sort("date").Limit(limit).All(&items)
	if err != nil {
		panic(err)
	}

	// set items to work channel
	for _, value := range items {
		work := tm.WorkRequest{Id: value.Id, Data: value}
		tm.NewWork(work)
	}

	// TODO delete this
	fmt.Printf("\n%d tasks added\n", len(items))
}
