package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/utils"
)

func (e mainEnv) userappNew(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	appName := strings.ToLower(ps.ByName("appname"))
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditAppEvent("create user app record by "+mode, identity, appName, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	if utils.CheckValidApp(appName) == false {
		utils.ReturnError(w, r, "bad appname", 405, nil, event)
		return
	}
	if e.db.store.ValidateNewApp("app_"+appName) == false {
		utils.ReturnError(w, r, "db limitation", 405, nil, event)
		return
	}

	data, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(data) == 0 {
		utils.ReturnError(w, r, "empty body", 405, nil, event)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	_, err = e.db.createAppRecord(jsonData, userTOKEN, appName, event, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	utils.ReturnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userappChange(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	appName := strings.ToLower(ps.ByName("appname"))
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditAppEvent("change user app record by "+mode, identity, appName, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	authResult := e.EnforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if utils.CheckValidApp(appName) == false {
		utils.ReturnError(w, r, "bad appname", 405, nil, event)
		return
	}
	jsonData, err := utils.GetJSONPostData(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if jsonData == nil {
		utils.ReturnError(w, r, "empty body", 405, nil, event)
		return
	}
	// make sure userapp exists
	resultJSON, err := e.db.getUserApp(userTOKEN, appName, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		utils.ReturnError(w, r, "not found", 405, nil, event)
		return
	}
	if authResult != "login" {
		_, err = e.db.updateAppRecord(jsonData, userTOKEN, appName, event, e.conf)
		if err != nil {
			utils.ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		utils.ReturnUUID(w, userTOKEN)
		return
	}
	if e.conf.SelfService.AppRecordChange != nil {
		for _, name := range e.conf.SelfService.AppRecordChange {
			if utils.StringPatternMatch(strings.ToLower(name), appName) {
				_, err = e.db.updateAppRecord(jsonData, userTOKEN, appName, event, e.conf)
				if err != nil {
					utils.ReturnError(w, r, "internal error", 405, err, event)
					return
				}
				utils.ReturnUUID(w, userTOKEN)
				return
			}
		}
	}
	rtoken, rstatus, err := e.db.saveUserRequest("change-app-data", userTOKEN, appName, "", jsonData, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
}

func (e mainEnv) userappList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditEvent("get user app list by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	result, err := e.db.listUserApps(userTOKEN, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":"%s","apps":%s}`, userTOKEN, result)
}

func (e mainEnv) userappGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	appName := strings.ToLower(ps.ByName("appname"))
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditAppEvent("get user app record by "+mode, identity, appName, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	if utils.CheckValidApp(appName) == false {
		utils.ReturnError(w, r, "bad appname", 405, nil, event)
		return
	}
	resultJSON, err := e.db.getUserApp(userTOKEN, appName, e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		utils.ReturnError(w, r, "not found", 405, nil, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s","data":%s}`, userTOKEN, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) userappDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	appName := strings.ToLower(ps.ByName("appname"))
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditAppEvent("delete user app record by "+mode, identity, appName, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	userTOKEN := e.loadUserToken(w, r, mode, identity, event)
	if userTOKEN == "" {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	if utils.CheckValidApp(appName) == false {
		utils.ReturnError(w, r, "bad appname", 405, nil, event)
		return
	}

	e.db.deleteUserApp(userTOKEN, appName, e.conf)

	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s"}`, userTOKEN)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) appList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAuth(w, r, nil) == "" {
		return
	}
	result, err := e.db.listAllApps(e.conf)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","apps":%s}`, result)
}
