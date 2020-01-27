package main

import (
	"encoding/json"
	"log"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type requestEvent struct {
	// time for update?
	Creationtime int32  `json:"creationtime"`
	When         int32  `json:"when"`
	Token        string `json:"token"`
	App          string `json:"app,omitempty"`
	Action       string `json:"action"`
	Status       string `json:"status"`
	Change       string `json:"change,omitempty"`
	Rtoken       string `json:"rtoken"`
}

func (dbobj dbcon) saveUserRequest(action string, token string, app string, change string) (string, error) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	rtoken, _ := uuid.GenerateUUID()
	bdoc["when"] = now
	bdoc["token"] = token
	bdoc["action"] = action
	bdoc["rtoken"] = rtoken
	bdoc["creationtime"] = now
	bdoc["status"] = "open"
	if len(change) > 0 {
		bdoc["change"] = change
	}
	if len(app) > 0 {
		bdoc["app"] = app
	}
	_, err := dbobj.createRecord(TblName.Requests, &bdoc)
	return rtoken, err
}

func (dbobj dbcon) getRequests(status string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.countRecords(TblName.Requests, "status", status)
	if err != nil {
		return nil, 0, err
	}
	var results []bson.M
	records, err := dbobj.getList(TblName.Requests, "status", status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, element := range records {
		element["more"] = false
		if _, ok := element["change"]; ok {
			element["more"] = true
			element["change"] = ""
		}
		results = append(results, element)
	}

	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) getRequest(rtoken string) (bson.M, error) {
	//var results []*auditEvent
	record, err := dbobj.getRecord(TblName.Requests, "rtoken", rtoken)
	if err != nil {
		return record, err
	}
	if len(record) == 0 {
		return record, err
	}
	//fmt.Printf("request record: %s\n", record)
	userTOKEN := ""
	change := ""
	if value, ok := record["token"]; ok {
		userTOKEN = value.(string)
	}
	if value, ok := record["change"]; ok {
		change = value.(string)
	}
	//recBson := bson.M{}
	if len(change) > 0 {
		change2, _ := dbobj.userDecrypt(userTOKEN, change)
		log.Printf("change: %s", change2)
		record["change"] = change2
	}
	return record, nil
}

func (dbobj dbcon) updateRequestStatus(rtoken string, status string) {
	bdoc := bson.M{}
	bdoc["status"] = status
	//fmt.Printf("op json: %s\n", update)
	dbobj.updateRecord(TblName.Requests, "rtoken", rtoken, &bdoc)
}
