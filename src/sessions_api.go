package main

import (
	"encoding/json"
	"fmt"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

func (e mainEnv) createSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	var event *auditEvent
	defer func() {
		if event != nil {
			event.submit(e.db)
		}
	}()
	if enforceUUID(w, session, event) == false {
		//returnError(w, r, "bad session format", nil, event)
		return
	}
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	expiration := e.conf.Policy.MaxSessionRetentionPeriod
	parsedData, err := getJSONPost(r, e.conf.Sms.DefaultCountry)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(parsedData.jsonData) == 0 {
		returnError(w, r, "empty request body", 405, nil, event)
		return
	}
	var userBson bson.M
	if len(parsedData.loginIdx) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("login", parsedData.loginIdx, e.conf)
	} else if len(parsedData.emailIdx) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("email", parsedData.emailIdx, e.conf)
	} else if len(parsedData.phoneIdx) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("phone", parsedData.phoneIdx, e.conf)
	} else if len(parsedData.customIdx) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("custom", parsedData.customIdx, e.conf)
	} else if len(parsedData.token) > 0 {
		userBson, err = e.db.lookupUserRecord(parsedData.token)
	}
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	userTOKEN := ""
	if userBson != nil {
		event = audit("create session", session, "session", session)
		userTOKEN = userBson["token"].(string)
		event.Record = userTOKEN
	}
	session, err = e.db.createSessionRecord(session, userTOKEN, expiration, parsedData.jsonData)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s"}`, session)
}

func (e mainEnv) deleteSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	event := audit("delete session", session, "session", session)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, session, event) == false {
		//returnError(w, r, "bad session format", nil, event)
		return
	}
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	e.db.deleteSession(session)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func (e mainEnv) newUserSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("create user session", identity, mode, identity)
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
		userBson, _ = e.db.lookupUserRecord(identity)
	} else {
		userBson, _ = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(records) == 0 {
		returnError(w, r, "empty body", 405, nil, event)
		return
	}
	expirationStr := getStringValue(records["expiration"])
	expiration := setExpiration(e.conf.Policy.MaxSessionRetentionPeriod, expirationStr)
	jsonData, err := json.Marshal(records)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	sessionUUID, err := uuid.GenerateUUID()
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	sessionID, err := e.db.createSessionRecord(sessionUUID, userTOKEN, expiration, jsonData)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s"}`, sessionID)
	return
}

func (e mainEnv) getUserSessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("get all user sessions", identity, mode, identity)
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
		userBson, _ = e.db.lookupUserRecord(identity)
	} else {
		// TODO: decode url in code!
		userBson, _ = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	e.db.store.DeleteExpired(storage.TblName.Sessions, "token", userTOKEN)
	args := r.URL.Query()
	var offset int32
	var limit int32 = 10
	if value, ok := args["offset"]; ok {
		offset = atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = atoi(value[0])
	}
	records, count, err := e.db.getUserSessionsByToken(userTOKEN, offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	data := strings.Join(records, ",")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","total":%d,"rows":[%s]}`, count, data)
	return
}

func (e mainEnv) getSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	var event *auditEvent
	defer func() {
		if event != nil {
			event.submit(e.db)
		}
	}()
	when, record, userTOKEN, err := e.db.getSession(session)
	if err != nil {
		returnError(w, r, err.Error(), 405, err, event)
		return
	}
	if len(userTOKEN) > 0 {
		event = audit("get session", session, "session", session)
		event.Record = userTOKEN
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s","when":%d,"data":%s}`, session, when, record)
	return
}
