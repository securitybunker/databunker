package main

import (
	"fmt"
	"net/http"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
)

func (e mainEnv) expUsers() error {
	records, err := e.db.store.GetExpiring(storage.TblName.Users, "expstatus", "wait")
	for _, rec := range records {
		userTOKEN := utils.GetUuidString(rec["token"])
		resultJSON, userBSON, _ := e.db.getUser(userTOKEN)
		if resultJSON != nil {
			email := utils.GetStringValue(userBSON["email"])
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
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")

	event := audit.CreateAuditEvent("get expiration status by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, userBSON, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}

	expirationDate := utils.GetIntValue(userBSON["endtime"])
	expirationStatus := utils.GetStringValue(userBSON["expstatus"])
	expirationToken := utils.GetStringValue(userBSON["exptoken"])
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
	event := audit.CreateAuditEvent("clear user expiration by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, _, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}

	status := ""
	err = e.db.updateUserExpStatus(userTOKEN, status)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
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
	event := audit.CreateAuditEvent("retain user data by exptoken", identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	if utils.EnforceUUID(w, identity, event) == false {
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
	if userBson == nil || err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
		return
	}
	userTOKEN := utils.GetUuidString(userBson["token"])
	event.Record = userTOKEN
	status := "retain"
	err = e.db.updateUserExpStatus(userTOKEN, status)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (e mainEnv) expDeleteData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("exptoken")
	mode := "exptoken"
	event := audit.CreateAuditEvent("delete user data by exptoken", identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	if utils.EnforceUUID(w, identity, event) == false {
		return
	}
	userJSON, userTOKEN, userBSON, err := e.db.getUserByIndex(identity, mode, e.conf)
	if userJSON == nil || err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
		return
	}
	event.Record = userTOKEN
	email := utils.GetStringValue(userBSON["email"])
	if len(email) > 0 {
		e.globalUserDelete(email)
	}
	_, err = e.db.deleteUserRecord(userJSON, userTOKEN, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
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
	event := audit.CreateAuditEvent("initiate user record expiration by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, _, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	//only admin should do this request?
	authResult := e.EnforceAdmin(w, r, event)
	if authResult == "" {
		return
	}

	postData, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	expirationStr := utils.GetStringValue(postData["expiration"])
	expiration := utils.SetExpiration(e.conf.Policy.MaxUserRetentionPeriod, expirationStr)
	endtime, _ := utils.ParseExpiration(expiration)
	status := utils.GetStringValue(postData["status"])
	if len(status) == 0 {
		status = "wait"
	}
	expToken, err := uuid.GenerateUUID()
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
	}
	err = e.db.initiateUserExpiration(userTOKEN, endtime, status, expToken)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","exptoken":"%s"}`, expToken)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}
