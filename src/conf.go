package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) setupConfRouter(router *httprouter.Router) *httprouter.Router {
	router.GET("/v1/sys/configuration", e.configurationDump)
	router.GET("/v1/sys/uiconfiguration", e.uiConfigurationDump)
	router.GET("/v1/sys/cookiesettings", e.cookieSettings)
	return router
}

func (e mainEnv) cookieSettings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	resultJSON, scriptsJSON, _, err := e.db.getLegalBasisCookieConf()
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	resultUIConfJSON, _ := json.Marshal(e.conf.UI)
	finalJSON := fmt.Sprintf(`{"status":"ok","ui":%s,"rows":%s,"scripts":%s}`, resultUIConfJSON, resultJSON, scriptsJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) configurationDump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	resultJSON, _ := json.Marshal(e.conf)
	finalJSON := fmt.Sprintf(`{"status":"ok","configuration":%s}`, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

// UI configuration dump API call.
func (e mainEnv) uiConfigurationDump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if len(e.conf.Notification.MagicSyncURL) != 0 &&
		len(e.conf.Notification.MagicSyncToken) != 0 {
		e.conf.UI.MagicLookup = true
	} else {
		e.conf.UI.MagicLookup = false
	}
	resultJSON, _ := json.Marshal(e.conf.UI)
	finalJSON := fmt.Sprintf(`{"status":"ok","ui":%s}`, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) globalUserDelete(email string) {
	// not implemented
}

func (dbobj dbcon) GetTenantAdmin(cfg Config) string {
	return cfg.Generic.AdminEmail
}

func (e mainEnv) pluginUserDelete(pluginid string, userTOKEN string) {
	// not implemented
}

func (e mainEnv) pluginUserLookup(email string) string {
	// not implemented
	return ""
}

func (dbobj dbcon) GlobalUserChangeEmail(oldEmail string, newEmail string) {
	// not implemented
}

func (dbobj dbcon) GetCode() []byte {
	code := dbobj.hash[4:12]
	return code
}
