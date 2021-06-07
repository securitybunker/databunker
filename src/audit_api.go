package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) getAuditEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit("view audit events", userTOKEN, "token", userTOKEN)
	defer func() { event.submit(e.db) }()
	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	var offset int32
	var limit int32 = 10
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = atoi(value[0])
	}
	resultJSON, counter, err := e.db.getAuditEvents(userTOKEN, offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	//fmt.Printf("Total count of events: %d\n", counter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getAdminAuditEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	var offset int32
	var limit int32 = 10
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = atoi(value[0])
	}
	resultJSON, counter, err := e.db.getAdminAuditEvents(offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	//fmt.Printf("Total count of events: %d\n", counter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getAuditEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	atoken := ps.ByName("atoken")
	event := audit("view audit event", atoken, "token", atoken)
	defer func() { event.submit(e.db) }()
	//fmt.Println("error code")
	if enforceUUID(w, atoken, event) == false {
		return
	}
	userTOKEN, resultJSON, err := e.db.getAuditEvent(atoken)
	log.Printf("extracted user token: %s", userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	event.Record = userTOKEN
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","event":%s}`, resultJSON)
	w.Write([]byte(str))
}
