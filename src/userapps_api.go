package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) userappNew(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := strings.ToLower(ps.ByName("appname"))
	event := auditApp("create user app record", userTOKEN, appName, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}
	if e.db.store.ValidateNewApp("app_"+appName) == false {
		returnError(w, r, "db limitation", 405, nil, event)
		return
	}

	data, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(data) == 0 {
		returnError(w, r, "empty body", 405, nil, event)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	_, err = e.db.createAppRecord(jsonData, userTOKEN, appName, event, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userappChange(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := strings.ToLower(ps.ByName("appname"))
	event := auditApp("change user app record", userTOKEN, appName, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}
	jsonData, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if jsonData == nil {
		returnError(w, r, "empty body", 405, nil, event)
		return
	}
	// make sure userapp exists
	resultJSON, err := e.db.getUserApp(userTOKEN, appName, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, nil, event)
		return
	}
	if authResult != "login" {
		_, err = e.db.updateAppRecord(jsonData, userTOKEN, appName, event, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		returnUUID(w, userTOKEN)
		return
	}
	if e.conf.SelfService.AppRecordChange != nil {
		for _, name := range e.conf.SelfService.AppRecordChange {
			if stringPatternMatch(strings.ToLower(name), appName) {
				_, err = e.db.updateAppRecord(jsonData, userTOKEN, appName, event, e.conf)
				if err != nil {
					returnError(w, r, "internal error", 405, err, event)
					return
				}
				returnUUID(w, userTOKEN)
				return
			}
		}
	}
	rtoken, rstatus, err := e.db.saveUserRequest("change-app-data", userTOKEN, appName, "", jsonData, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
}

func (e mainEnv) userappList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit("get user app list", userTOKEN, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	result, err := e.db.listUserApps(userTOKEN, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":"%s","apps":%s}`, userTOKEN, result)
}

func (e mainEnv) userappGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := strings.ToLower(ps.ByName("appname"))
	event := auditApp("get user app record", userTOKEN, appName, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}
	resultJSON, err := e.db.getUserApp(userTOKEN, appName, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, nil, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s","data":%s}`, userTOKEN, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) userappDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := strings.ToLower(ps.ByName("appname"))
	event := auditApp("delete user app record", userTOKEN, appName, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}

	e.db.deleteUserApp(userTOKEN, appName, e.conf)

	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s"}`, userTOKEN)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) appList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	result, err := e.db.listAllApps(e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","apps":%s}`, result)
}
