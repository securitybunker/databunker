package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/utils"
)

func (e mainEnv) EnforceAuth(w http.ResponseWriter, r *http.Request, event *audit.AuditEvent) string {
	/*
		for key, value := range r.Header {
			fmt.Printf("%s => %s\n", key, value)
		}
	*/
	if token, ok := r.Header["X-Bunker-Token"]; ok {
		authResult, err := e.db.checkUserAuthXToken(token[0])
		//fmt.Printf("error in auth? error %s - %s\n", err, token[0])
		if err == nil {
			if event != nil {
				event.Identity = authResult.name
				if authResult.ttype == "login" && authResult.token == event.Record {
					return authResult.ttype
				}
			}
			if len(authResult.ttype) > 0 && authResult.ttype != "login" {
				return authResult.ttype
			}
		}
		/*
			if e.db.checkXtoken(token[0]) == true {
				if event != nil {
					event.Identity = "admin"
				}
				return true
			}
		*/
	}
	log.Printf("403 Access denied\n")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Access denied"))
	if event != nil {
		event.Status = "error"
		event.Msg = "access denied"
	}
	return ""
}

func (e mainEnv) EnforceAdmin(w http.ResponseWriter, r *http.Request, event *audit.AuditEvent) string {
	if token, ok := r.Header["X-Bunker-Token"]; ok {
		authResult, err := e.db.checkUserAuthXToken(token[0])
		//fmt.Printf("error in auth? error %s - %s\n", err, token[0])
		if err == nil {
			if event != nil {
				event.Identity = authResult.name
			}
			if len(authResult.ttype) > 0 && authResult.ttype != "login" {
				return authResult.ttype
			}
		}
	}
	log.Printf("403 Access denied\n")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Access denied"))
	return ""
}

func (e mainEnv) getAuditEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit.CreateAuditEvent("view audit events", userTOKEN, "token", userTOKEN)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	if utils.EnforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.EnforceAuth(w, r, event) == "" {
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
	resultJSON, counter, err := e.db.getAuditEvents(userTOKEN, offset, limit)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	//fmt.Printf("Total count of events: %d\n", counter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getAdminAuditEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAdmin(w, r, nil) == "" {
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
	resultJSON, counter, err := e.db.getAdminAuditEvents(offset, limit)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
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
	event := audit.CreateAuditEvent("view audit event", atoken, "token", atoken)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	//fmt.Println("error code")
	if utils.EnforceUUID(w, atoken, event) == false {
		return
	}
	userTOKEN, resultJSON, err := e.db.getAuditEvent(atoken)
	log.Printf("extracted user token: %s", userTOKEN)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	event.Record = userTOKEN
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","event":%s}`, resultJSON)
	w.Write([]byte(str))
}
