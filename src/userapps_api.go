package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) userappNew(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := ps.ByName("appname")
	event := auditApp("create user app record", userTOKEN, appName)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}
	if e.db.validateNewApp("app_"+appName) == false {
		returnError(w, r, "db limitation", 405, nil, event)
		return
	}

	data, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	_, err = e.db.createAppRecord(jsonData, userTOKEN, appName, event)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userappChange(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	appName := ps.ByName("appname")
	event := auditApp("change user app record", userTOKEN, appName)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}

	data, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	_, err = e.db.updateAppRecord(jsonData, userTOKEN, appName, event)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userappList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit("get user app list", userTOKEN, "token", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
		return
	}
	result, err := e.db.listUserApps(userTOKEN)
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
	appName := ps.ByName("appname")
	event := auditApp("get user app record", userTOKEN, appName)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
		return
	}
	if isValidApp(appName) == false {
		returnError(w, r, "bad appname", 405, nil, event)
		return
	}

	resultJSON, err := e.db.getUserApp(userTOKEN, appName)
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

func (e mainEnv) appList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Printf("/APPLIST\n")
	if e.enforceAuth(w, r, nil) == false {
		return
	}
	result, err := e.db.listAllApps()
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","apps":%s}`, result)
}
