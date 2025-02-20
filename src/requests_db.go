package main

import (
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) saveUserRequest(action, token string, userBSON map[string]interface{}, app, brief string, change []byte, cfg Config) (string, string, error) {
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
		return utils.GetUuidString(record["rtoken"]), "request-exists", nil
	}
	rtoken, _ := uuid.GenerateUUID()
	bdoc["when"] = now
	bdoc["rtoken"] = rtoken
	bdoc["creationtime"] = now
	if change != nil {
		encodedStr, err := dbobj.userEncrypt(userBSON, change)
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
		element["token"] = utils.GetUuidString(element["token"])
		element["rtoken"] = utils.GetUuidString(element["rtoken"])
	}

	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
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
		element["token"] = userTOKEN
		element["rtoken"] = utils.GetUuidString(element["rtoken"])
		results = append(results, element)
	}

	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	return resultJSON, count, nil
}

func (dbobj dbcon) getRequest(rtoken string) (map[string]interface{}, error) {
	record, err := dbobj.store.GetRecord(storage.TblName.Requests, "rtoken", rtoken)
	if err != nil {
		return record, err
	}
	if len(record) == 0 {
		return record, err
	}
	//fmt.Printf("request record: %s\n", record)
	userTOKEN := ""
	
	userTOKEN = utils.GetUuidString(record["token"])
	record["token"] = userTOKEN
	record["rtoken"] = utils.GetUuidString(record["rtoken"])
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
