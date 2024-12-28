package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/utils"
)

// This function retrieves all requests that require admin approval. This function supports result pager.
func (e mainEnv) getUserRequests(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAuth(w, r, nil) == "" {
		return
	}
	var offset int32
	var limit int32 = 10
	status := "open"
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = utils.Atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = utils.Atoi(value[0])
	}
	if value, ok := args["status"]; ok {
		status = value[0]
	}
	resultJSON, counter, err := e.db.getRequests(status, offset, limit)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

// Get list of requests for specific user
func (e mainEnv) getCustomUserRequests(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditEvent("get user privacy requests", identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	var offset int32
	var limit int32 = 10
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = utils.Atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = utils.Atoi(value[0])
	}
	resultJSON, counter, err := e.db.getUserRequests(userTOKEN, offset, limit)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getUserRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	request := ps.ByName("request")
	event := audit.CreateAuditEvent("get user request by request token", request, "request", request)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if utils.EnforceUUID(w, request, event) == false {
		return
	}
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	if len(requestInfo) == 0 {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	var resultJSON []byte
	action := utils.GetStringValue(requestInfo["action"])
	userTOKEN := utils.GetStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	if e.EnforceAdmin(w, r, event) == "" {
		return
	}
	change := utils.GetStringValue(requestInfo["change"])
	appName := utils.GetStringValue(requestInfo["app"])
	brief := utils.GetStringValue(requestInfo["brief"])
	if strings.HasPrefix(action, "plugin") {
		brief = ""
	}
	if len(appName) > 0 {
		resultJSON, err = e.db.getUserApp(userTOKEN, appName, e.conf)
	} else if len(brief) > 0 {
		resultJSON, err = e.db.viewAgreementRecord(userTOKEN, brief)
	} else {
		resultJSON, err = e.db.getUserJSON(userTOKEN)
	}
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	//fmt.Printf("Full json: %s\n", resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	var str string

	if len(appName) > 0 {
		str = fmt.Sprintf(`"status":"ok","app":"%s"`, appName)
	} else {
		str = fmt.Sprintf(`"status":"ok"`)
	}
	if len(resultJSON) > 0 {
		str = fmt.Sprintf(`%s,"original":%s`, str, resultJSON)
	}
	if len(change) > 0 {
		str = fmt.Sprintf(`%s,"change":%s`, str, change)
	}
	str = fmt.Sprintf(`{%s}`, str)
	//fmt.Printf("result: %s\n", str)
	w.Write([]byte(str))
}

func (e mainEnv) approveUserRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	request := ps.ByName("request")
	event := audit.CreateAuditEvent("approve user request", request, "request", request)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if utils.EnforceUUID(w, request, event) == false {
		return
	}
	authResult := e.EnforceAdmin(w, r, event)
	if authResult == "" {
		return
	}
	records, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	reason := utils.GetStringValue(records["reason"])
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if len(requestInfo) == 0 {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	userTOKEN := utils.GetStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	action := utils.GetStringValue(requestInfo["action"])
	status := utils.GetStringValue(requestInfo["status"])
	if status != "open" {
		utils.ReturnError(w, r, "wrong status: "+status, 405, err, event)
		return
	}
	userJSON, userBSON, err := e.db.getUser(userTOKEN)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if userJSON == nil {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	if action == "forget-me" {
		email := utils.GetStringValue(userBSON["email"])
		if len(email) > 0 {
			e.globalUserDelete(email)
		}
		result, err := e.db.deleteUserRecord(userJSON, userTOKEN, e.conf)
		if result == false || err != nil {
			// user deleted
			event.Status = "failed"
			event.Msg = "failed to delete"
		}
		if err != nil {
			utils.ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		notifyURL := e.conf.Notification.NotificationURL
		notifyForgetMe(notifyURL, userJSON, "token", userTOKEN)
	} else if action == "change-profile" {
		oldJSON, newJSON, lookupErr, err := e.db.updateUserRecord(requestInfo["change"].([]uint8), userTOKEN, userBSON, event, e.conf)
		if lookupErr {
			utils.ReturnError(w, r, "internal error", 405, errors.New("not found"), event)
			return
		}
		if err != nil {
			utils.ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		utils.ReturnUUID(w, userTOKEN)
		notifyURL := e.conf.Notification.NotificationURL
		notifyProfileChange(notifyURL, oldJSON, newJSON, "token", userTOKEN)
	} else if action == "change-app-data" {
		app := requestInfo["app"].(string)
		_, err = e.db.updateAppRecord(requestInfo["change"].([]uint8), userTOKEN, app, event, e.conf)
		if err != nil {
			utils.ReturnError(w, r, "internal error", 405, err, event)
			return
		}
	} else if action == "agreement-withdraw" {
		brief := requestInfo["brief"].(string)
		mode := "token"
		lastmodifiedby := "admin"
		e.db.withdrawAgreement(userTOKEN, brief, mode, userTOKEN, lastmodifiedby)
	} else if action == "plugin-delete" {
		pluginid := requestInfo["brief"].(string)
		e.pluginUserDelete(pluginid, userTOKEN)
	}
	e.db.updateRequestStatus(request, "approved", reason)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}

func (e mainEnv) cancelUserRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	request := ps.ByName("request")
	event := audit.CreateAuditEvent("cancel user request", request, "request", request)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if utils.EnforceUUID(w, request, event) == false {
		return
	}
	records, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	reason := utils.GetStringValue(records["reason"])
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if len(requestInfo) == 0 {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	userTOKEN := utils.GetStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	authResult := e.EnforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if requestInfo["status"].(string) != "open" {
		utils.ReturnError(w, r, "wrong status: "+requestInfo["status"].(string), 405, err, event)
		return
	}
	resultJSON, err := e.db.getUserJSON(userTOKEN)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		utils.ReturnError(w, r, "not found", 405, err, event)
		return
	}
	if len(reason) == 0 && authResult == "login" {
		reason = "user operation"
	}
	e.db.updateRequestStatus(request, "canceled", reason)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}
