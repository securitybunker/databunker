package main

import (
	"fmt"
	"net/http"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) expUsers() error {
	records, err := e.db.store.GetExpiring(storage.TblName.Users, "expstatus", "wait")
	for _, rec := range records {
		userTOKEN := rec["token"].(string)
		resultJSON, userBSON, _ := e.db.getUser(userTOKEN)
		if resultJSON != nil {
			email := getStringValue(userBSON["email"])
			if len(email) > 0 {
				e.globalUserDelete(email)
			}
			e.db.deleteUserRecord(resultJSON, userTOKEN, e.conf)
			e.db.updateUserExpStatus(userTOKEN, "expired")
		}
	}
	return err
}

func (e mainEnv) expGetStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("get expiration status by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := identity
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(identity)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
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
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("clear user expiration by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := identity
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(identity)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
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
	identity := ps.ByName("exptoken")
	mode := "exptoken"
	event := audit("retain user data by exptoken", identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, identity, event) == false {
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
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
	identity := ps.ByName("exptoken")
	mode := "exptoken"
	event := audit("delete user data by exptoken", identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, identity, event) == false {
		return
	}
	userJSON, userTOKEN, userBSON, err := e.db.getUserByIndex(identity, mode, e.conf)
	if userJSON == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	event.Record = userTOKEN
	email := getStringValue(userBSON["email"])
	if len(email) > 0 {
		e.globalUserDelete(email)
	}
	_, err = e.db.deleteUserRecord(userJSON, userTOKEN, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	e.db.updateUserExpStatus(userTOKEN, "expired")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (e mainEnv) expStart(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("initiate user record expiration by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	if e.enforceAdmin(w, r) == "" {
		return
	}
	userTOKEN := identity
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, err = e.db.lookupUserRecord(identity)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	expirationStr := getStringValue(records["expiration"])
	expiration := setExpiration(e.conf.Policy.MaxUserRetentionPeriod, expirationStr)
	endtime, _ := parseExpiration(expiration)
	status := getStringValue(records["status"])
	if len(status) == 0 {
		status = "wait"
	}
	expToken, err := uuid.GenerateUUID()
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
	}
	err = e.db.initiateUserExpiration(userTOKEN, endtime, status, expToken)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","exptoken":"%s"}`, expToken)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}
