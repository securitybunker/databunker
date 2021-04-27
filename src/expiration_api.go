package main

import (
	"fmt"
	"net/http"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) expGetStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("get expiration status by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := address
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(address)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	expirationDate := getIntValue(userBson["endtime"])
	expirationStatus := getStringValue(userBson["expstatus"])
	expirationToken := getStringValue(userBson["exptoken"])
	finalJSON := fmt.Sprintf(`{"status":"ok","exptime":%d,"expstatus":"%s","exptoken":"%s"}`,
		expirationDate, expirationStatus, expirationToken)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) expCancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("clear user expiration by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := address
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(address)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	status := ""
	err = e.db.updateUserExpStatus(userTOKEN, status)
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	finalJSON := `{"status":"ok"}`
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) expRetainData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("exptoken")
	mode := "exptoken"
	event := audit("retain user data by exptoken", address, mode, address)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, address, event) == false {
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	userTOKEN := userBson["token"].(string)
	event.Record = userTOKEN
	status := "retain"
	err = e.db.updateUserExpStatus(userTOKEN, status)
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (e mainEnv) expDeleteData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("exptoken")
	mode := "exptoken"
	event := audit("delete user data by exptoken", address, mode, address)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, address, event) == false {
		return
	}
	resultJSON, userTOKEN, err := e.db.getUserJsonByIndex(address, mode, e.conf)
	if resultJSON == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	event.Record = userTOKEN
	e.globalUserDelete(userTOKEN)
	_, err = e.db.deleteUserRecord(resultJSON, userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (e mainEnv) expInitiate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("initiate user record expiration by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	if e.enforceAdmin(w, r) == "" {
		return
	}
	userTOKEN := address
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(address)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	records, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	expirationStr := getStringValue(records["expiration"])
	expiration := setExpiration(e.conf.Policy.MaxUserRetentionPeriod, expirationStr)
	status := getStringValue(records["status"])
	if len(status) == 0 {
		status = "wait"
	}
	expToken, err := uuid.GenerateUUID()
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
	}
	err = e.db.initiateUserExpiration(userTOKEN, expiration, status, expToken)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","exptoken":"%s"}`, expToken)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

