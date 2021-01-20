package main

import (
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type requestEvent struct {
	// time for update?
	Creationtime int32  `json:"creationtime"`
	When         int32  `json:"when"`
	Token        string `json:"token"`
	App          string `json:"app,omitempty"`
	Brief        string `json:"brief,omitempty"`
	Action       string `json:"action"`
	Status       string `json:"status"`
	Change       string `json:"change,omitempty"`
	Rtoken       string `json:"rtoken"`
	Reason       string `json:"reason"`
}

func (dbobj dbcon) saveUserRequest(action string, token string, app string, brief string, change []byte, cfg Config) (string, string, error) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	bdoc["token"] = token
	bdoc["action"] = action
	bdoc["status"] = "open"
	if len(app) > 0 {
		bdoc["app"] = app
	}
	if len(brief) > 0 {
		bdoc["brief"] = brief
	}
	record, err := dbobj.store.LookupRecord(storage.TblName.Requests, bdoc)
	if record != nil {
		fmt.Printf("This record already exists.\n")
		return record["rtoken"].(string), "request-exists", nil
	}
	rtoken, _ := uuid.GenerateUUID()
	bdoc["when"] = now
	bdoc["rtoken"] = rtoken
	bdoc["creationtime"] = now
	if change != nil {
		encodedStr, err := dbobj.userEncrypt(token, change)
		if err != nil {
			return "", "", err
		}
		bdoc["change"] = encodedStr
	}
	_, err = dbobj.store.CreateRecord(storage.TblName.Requests, &bdoc)
	if err != nil {
		adminEmail := dbobj.GetTenantAdmin(cfg)
		if len(adminEmail) > 0 {
			go adminEmailAlert(action, adminEmail, cfg)
		}
	}
	return rtoken, "request-created", err
}

func (dbobj dbcon) getRequests(status string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.store.CountRecords(storage.TblName.Requests, "status", status)
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return []byte("[]"), 0, err
	}
	var results []bson.M
	records, err := dbobj.store.GetList(storage.TblName.Requests, "status", status, offset, limit, "when")
	if err != nil {
		return nil, 0, err
	}
	for _, element := range records {
		element["more"] = false
		if _, ok := element["change"]; ok {
			element["more"] = true
			delete(element, "change")
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

func (dbobj dbcon) getUserRequests(userTOKEN string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.store.CountRecords(storage.TblName.Requests, "token", userTOKEN)
	if err != nil {
		return nil, 0, err
	}
	var results []bson.M
	records, err := dbobj.store.GetList(storage.TblName.Requests, "token", userTOKEN, offset, limit, "")
	if err != nil {
		return nil, 0, err
	}
	for _, element := range records {
		element["more"] = false
		if _, ok := element["change"]; ok {
			element["more"] = true
			delete(element, "change")
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
	record, err := dbobj.store.GetRecord(storage.TblName.Requests, "rtoken", rtoken)
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
		//log.Printf("change: %s", change2)
		record["change"] = change2
	}
	return record, nil
}

func (dbobj dbcon) updateRequestStatus(rtoken string, status string, reason string) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	bdoc["status"] = status
	bdoc["when"] = now
	if len(reason) > 0 {
		bdoc["reason"] = reason
	}
	//fmt.Printf("op json: %s\n", update)
	dbobj.store.UpdateRecord(storage.TblName.Requests, "rtoken", rtoken, &bdoc)
}
