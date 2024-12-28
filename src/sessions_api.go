package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) createSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	var event *audit.AuditEvent
	defer func() {
		if event != nil {
			SaveAuditEvent(event, e.db, e.conf)
		}
	}()
	if utils.EnforceUUID(w, session, event) == false {
		//utils.ReturnError(w, r, "bad session format", nil, event)
		return
	}
	if e.EnforceAdmin(w, r, event) == "" {
		return
	}
	records, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(records) == 0 {
		utils.ReturnError(w, r, "empty body", 405, nil, event)
		return
	}
	expirationStr := utils.GetStringValue(records["expiration"])
	expiration := utils.SetExpiration(e.conf.Policy.MaxSessionRetentionPeriod, expirationStr)
	log.Printf("Record expiration: %s", expiration)
	userToken := utils.GetStringValue(records["token"])
	userLogin := utils.GetStringValue(records["login"])
	userEmail := utils.GetStringValue(records["email"])
	userPhone := utils.GetStringValue(records["phone"])
	userCustomIdx := utils.GetStringValue(records["custom"])

	var userBson bson.M
	if len(userLogin) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("login", userLogin, e.conf)
	} else if len(userEmail) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("email", userEmail, e.conf)
	} else if len(userPhone) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("phone", userPhone, e.conf)
	} else if len(userCustomIdx) > 0 {
		userBson, err = e.db.lookupUserRecordByIndex("custom", userCustomIdx, e.conf)
	} else if len(userToken) > 0 {
		userBson, err = e.db.lookupUserRecord(userToken)
	}
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	userTOKEN := ""
	if userBson != nil {
		event = audit.CreateAuditEvent("create session", session, "session", session)
		userTOKEN = userBson["token"].(string)
		event.Record = userTOKEN
	}
	jsonData, err := json.Marshal(records)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	session, err = e.db.createSessionRecord(session, userTOKEN, expiration, jsonData)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s"}`, session)
}

func (e mainEnv) deleteSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	event := audit.CreateAuditEvent("delete session", session, "session", session)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	if utils.EnforceUUID(w, session, event) == false {
		//utils.ReturnError(w, r, "bad session format", nil, event)
		return
	}
	if e.EnforceAdmin(w, r, event) == "" {
		return
	}
	e.db.deleteSession(session)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

// the following function is currently not used
func (e mainEnv) newUserSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditEvent("create user session by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	records, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(records) == 0 {
		utils.ReturnError(w, r, "empty body", 405, nil, event)
		return
	}
	expirationStr := utils.GetStringValue(records["expiration"])
	expiration := utils.SetExpiration(e.conf.Policy.MaxSessionRetentionPeriod, expirationStr)
	log.Printf("Record expiration: %s", expiration)
	jsonData, err := json.Marshal(records)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	sessionUUID, err := uuid.GenerateUUID()
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	sessionID, err := e.db.createSessionRecord(sessionUUID, userTOKEN, expiration, jsonData)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s"}`, sessionID)
}

func (e mainEnv) getUserSessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditEvent("get all user sessions", identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	e.db.store.DeleteExpired(storage.TblName.Sessions, "token", userTOKEN)
	args := r.URL.Query()
	var offset int32
	var limit int32 = 10
	if value, ok := args["offset"]; ok {
		offset = utils.Atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = utils.Atoi(value[0])
	}
	records, count, err := e.db.getUserSessionsByToken(userTOKEN, offset, limit)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	data := strings.Join(records, ",")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","total":%d,"rows":[%s]}`, count, data)
}

func (e mainEnv) getSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := ps.ByName("session")
	event := audit.CreateAuditEvent("get session", session, "session", session)
	defer func() {
		if event != nil {
			SaveAuditEvent(event, e.db, e.conf)
		}
	}()
	when, record, userTOKEN, err := e.db.getSession(session)
	if len(userTOKEN) > 0 {
		event.Record = userTOKEN
		e.db.store.DeleteExpired(storage.TblName.Sessions, "token", userTOKEN)
	}
	if err != nil {
		utils.ReturnError(w, r, err.Error(), 405, err, event)
		return
	}

	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","session":"%s","when":%d,"data":%s}`, session, when, record)
}
