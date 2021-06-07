package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type processingActivity struct {
	Activity     string `json:"activity" structs:"activity"`
	Title        string `json:"title" structs:"title"`
	Script       string `json:"script,omitempty" structs:"script,omitempty"`
	Fulldesc     string `json:"fulldesc,omitempty" structs:"fulldesc,omitempty"`
	Legalbasis   string `json:"legalbasis,omitempty" structs:"legalbasis,omitempty"`
	Applicableto string `json:"applicableto,omitempty" structs:"applicableto,omitempty"`
	Creationtime int32  `json:"creationtime" structs:"creationtime"`
}

func (dbobj dbcon) createProcessingActivity(activity string, newactivity string, title string, script string, fulldesc string, legalbasis string, applicableto string) (bool, error) {
	bdoc := bson.M{}
	bdoc["title"] = title
	bdoc["script"] = script
	bdoc["fulldesc"] = fulldesc
	if len(legalbasis) > 0 {
		bdoc["legalbasis"] = legalbasis
	}
	bdoc["applicableto"] = applicableto
	raw, err := dbobj.store.GetRecord(storage.TblName.Processingactivities, "activity", activity)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return false, err
	}
	if raw != nil {
		if len(newactivity) > 0 && newactivity != activity {
			bdoc["activity"] = newactivity
		}
		_, err = dbobj.store.UpdateRecord(storage.TblName.Processingactivities, "activity", activity, &bdoc)
		return false, err
	}
	now := int32(time.Now().Unix())
	bdoc["activity"] = activity
	bdoc["creationtime"] = now
	_, err = dbobj.store.CreateRecord(storage.TblName.Processingactivities, &bdoc)
	if err != nil {
		fmt.Printf("error to insert record: %s\n", err)
		return false, err
	}
	return true, nil
}

func (dbobj dbcon) deleteProcessingActivity(activity string) (bool, error) {
	dbobj.store.DeleteRecord(storage.TblName.Processingactivities, "activity", activity)
	return true, nil
}

func (dbobj dbcon) linkProcessingActivity(activity string, brief string) (bool, error) {
	raw, err := dbobj.store.GetRecord(storage.TblName.Processingactivities, "activity", activity)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return false, err
	}
	if raw == nil {
		return false, errors.New("not found")
	}
	legalbasis := ""
	if value, ok := raw["legalbasis"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			legalbasis = value.(string)
		}
	}
	briefs := strings.Split(legalbasis, ",")
	if contains(briefs, brief) == true {
		// nothing to do here
		return false, nil
	}
	if len(legalbasis) > 0 {
		legalbasis = legalbasis + "," + brief
	} else {
		legalbasis = brief
	}
	bdoc := bson.M{}
	bdoc["legalbasis"] = legalbasis
	_, err = dbobj.store.UpdateRecord(storage.TblName.Processingactivities, "activity", activity, &bdoc)
	return true, err
}

func (dbobj dbcon) unlinkProcessingActivity(activity string, brief string) (bool, error) {
	raw, err := dbobj.store.GetRecord(storage.TblName.Processingactivities, "activity", activity)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return false, err
	}
	if raw == nil {
		return false, errors.New("not found")
	}
	legalbasis := ""
	if val, ok := raw["legalbasis"]; ok {
		if reflect.TypeOf(val) == reflect.TypeOf("string") {
			legalbasis = val.(string)
		}
	}
	briefs := strings.Split(legalbasis, ",")
	if contains(briefs, brief) == false {
		// nothing to do here
		return false, nil
	}
	legalbasis = ""
	found := false
	for _, value := range briefs {
		if value != brief {
			if len(legalbasis) > 0 {
				legalbasis = legalbasis + "," + value
			} else {
				legalbasis = value
			}
		} else {
			found = true
		}
	}
	if found == false {
		return true, nil
	}
	bdoc := bson.M{}
	bdoc["legalbasis"] = legalbasis
	_, err = dbobj.store.UpdateRecord(storage.TblName.Processingactivities, "activity", activity, &bdoc)
	return true, err
}

func (dbobj dbcon) unlinkProcessingActivityBrief(brief string) (bool, error) {
	records, err := dbobj.store.GetList0(storage.TblName.Processingactivities, 0, 0, "")
	if err != nil {
		return false, err
	}
	for _, record := range records {
		legalbasis := ""
		found := false
		if val, ok := record["legalbasis"]; ok {
			briefs := strings.Split(val.(string), ",")
			if len(briefs) > 0 {
				for _, value := range briefs {
					if value != brief {
						if len(legalbasis) > 0 {
							legalbasis = legalbasis + "," + value
						} else {
							legalbasis = value
						}
					} else {
						found = true
					}
				}
			}
		}
		if found == true {
			bdoc := bson.M{}
			bdoc["legalbasis"] = legalbasis
			dbobj.store.UpdateRecord(storage.TblName.Processingactivities, "activity", record["activity"].(string), &bdoc)
		}
	}
	return true, nil
}

func (dbobj dbcon) listProcessingActivities() ([]byte, int, error) {
	set := make(map[string]interface{})
	records0, err := dbobj.store.GetList0(storage.TblName.Legalbasis, 0, 0, "")
	for _, val := range records0 {
		set[val["brief"].(string)] = val
	}
	records, err := dbobj.store.GetList0(storage.TblName.Processingactivities, 0, 0, "")
	if err != nil {
		return nil, 0, err
	}
	count := len(records)
	if count == 0 {
		return []byte("[]"), 0, err
	}
	var results []interface{}
	for _, record := range records {
		var results0 []interface{}
		if record["legalbasis"] != nil {
			briefs := strings.Split(record["legalbasis"].(string), ",")
			if len(briefs) > 0 {
				for _, brief := range briefs {
					if value, ok := set[brief]; ok {
						results0 = append(results0, value)
					}
				}
			}
		}
		record["briefs"] = results0
		results = append(results, record)
	}
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}
