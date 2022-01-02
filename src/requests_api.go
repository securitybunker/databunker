package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) getUserRequests(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	var offset int32
	var limit int32 = 10
	status := "open"
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = atoi(value[0])
	}
	if value, ok := args["status"]; ok {
		status = value[0]
	}
	resultJSON, counter, err := e.db.getRequests(status, offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getCustomUserRequests(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("get user privacy requests", identity, mode, identity)
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
	var offset int32
	var limit int32 = 10
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = atoi(value[0])
	}
	resultJSON, counter, err := e.db.getUserRequests(userTOKEN, offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	fmt.Printf("Total count of custom user requests: %d\n", counter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getUserRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	request := ps.ByName("request")
	event := audit("get user request by request token", request, "request", request)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, request, event) == false {
		return
	}
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	if len(requestInfo) == 0 {
		returnError(w, r, "not found", 405, err, event)
		return
	}
	var resultJSON []byte
	action := getStringValue(requestInfo["action"])
	userTOKEN := getStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	change := getStringValue(requestInfo["change"])
	appName := getStringValue(requestInfo["app"])
	brief := getStringValue(requestInfo["brief"])
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
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, err, event)
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
	event := audit("approve user request", request, "request", request)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, request, event) == false {
		return
	}
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	reason := getStringValue(records["reason"])
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if len(requestInfo) == 0 {
		returnError(w, r, "not found", 405, err, event)
		return
	}
	userTOKEN := getStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	action := getStringValue(requestInfo["action"])
	status := getStringValue(requestInfo["status"])
	if status != "open" {
		returnError(w, r, "wrong status: "+status, 405, err, event)
		return
	}
	userJSON, userBSON, err := e.db.getUser(userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if userJSON == nil {
		returnError(w, r, "not found", 405, err, event)
		return
	}
	if action == "forget-me" {
		email := getStringValue(userBSON["email"])
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
                        returnError(w, r, "internal error", 405, err, event)
                        return
                }
		notifyURL := e.conf.Notification.NotificationURL
		notifyForgetMe(notifyURL, userJSON, "token", userTOKEN)
	} else if action == "change-profile" {
		oldJSON, newJSON, lookupErr, err := e.db.updateUserRecord(requestInfo["change"].([]uint8), userTOKEN, userBSON, event, e.conf)
		if lookupErr {
			returnError(w, r, "internal error", 405, errors.New("not found"), event)
			return
		}
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		returnUUID(w, userTOKEN)
		notifyURL := e.conf.Notification.NotificationURL
		notifyProfileChange(notifyURL, oldJSON, newJSON, "token", userTOKEN)
	} else if action == "change-app-data" {
		app := requestInfo["app"].(string)
		_, err = e.db.updateAppRecord(requestInfo["change"].([]uint8), userTOKEN, app, event, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
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
	event := audit("cancel user request", request, "request", request)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, request, event) == false {
		return
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	reason := getStringValue(records["reason"])
	requestInfo, err := e.db.getRequest(request)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if len(requestInfo) == 0 {
		returnError(w, r, "not found", 405, err, event)
		return
	}
	userTOKEN := getStringValue(requestInfo["token"])
	if len(userTOKEN) != 0 {
		event.Record = userTOKEN
	}
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if requestInfo["status"].(string) != "open" {
		returnError(w, r, "wrong status: "+requestInfo["status"].(string), 405, err, event)
		return
	}
	resultJSON, err := e.db.getUserJSON(userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, err, event)
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
