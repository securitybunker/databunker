package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/utils"
	"github.com/tidwall/gjson"
)

func (e mainEnv) sharedRecordCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit.CreateAuditEvent("create shareable record by "+mode, identity, "token", identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, _, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	postData, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	fields := utils.GetStringValue(postData["fields"])
	session := utils.GetStringValue(postData["session"])
	partner := utils.GetStringValue(postData["partner"])
	appName := utils.GetStringValue(postData["app"])
	expiration := e.conf.Policy.MaxShareableRecordRetentionPeriod

	if len(appName) > 0 {
		appName = strings.ToLower(appName)
		if utils.CheckValidApp(appName) == false {
			utils.ReturnError(w, r, "unknown app name", 405, nil, event)
		}
	}
	if value, ok := postData["expiration"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			expiration = utils.SetExpiration(e.conf.Policy.MaxShareableRecordRetentionPeriod, value.(string))
		} else {
			utils.ReturnError(w, r, "failed to parse expiration field", 405, err, event)
			return
		}
	}
	if len(expiration) == 0 {
		// using default expiration time for record
		expiration = "1m"
	}
	recordUUID, err := e.db.saveSharedRecord(userTOKEN, fields, expiration, session, appName, partner, e.conf)
	if err != nil {
		utils.ReturnError(w, r, err.Error(), 405, err, event)
		return
	}
	event.Record = userTOKEN
	event.Msg = "generated: " + recordUUID
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","record":%q}`, recordUUID)
}

func (e mainEnv) sharedRecordGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	record := ps.ByName("record")
	event := audit.CreateAuditEvent("get shareable record by token", record, "record", record)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if utils.EnforceUUID(w, record, event) == false {
		return
	}
	recordInfo, err := e.db.getSharedRecord(record)
	if err != nil {
		log.Printf("%d access denied for : %s\n", http.StatusForbidden, record)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}
	var resultJSON []byte
	if len(recordInfo.token) > 0 {
		event.Record = recordInfo.token
		event.App = recordInfo.appName
		log.Printf("fields to display: %s, user token: %s\n", recordInfo.fields, recordInfo.token)

		if len(recordInfo.appName) > 0 {
			if len(recordInfo.token) > 0 {
				userBSON, err := e.db.lookupUserRecord(recordInfo.token)
				if err != nil {
					utils.ReturnError(w, r, "internal error", 405, err, event)
					return
				}
				resultJSON, err = e.db.getUserApp(recordInfo.token, userBSON, recordInfo.appName, e.conf)
			}
		} else if len(recordInfo.session) > 0 {
			_, resultJSON, _, err = e.db.getSession(recordInfo.session)
		} else {
			resultJSON, err = e.db.getUserJSON(recordInfo.token)
		}
		if err != nil {
			utils.ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		if resultJSON == nil {
			utils.ReturnError(w, r, "not found", 405, err, event)
			return
		}
		//log.Printf("Full json: %s\n", resultJSON)
		if len(recordInfo.fields) > 0 {
			raw := make(map[string]interface{})
			//var newJSON json
			allFields := utils.ParseFields(recordInfo.fields)
			for _, f := range allFields {
				if f == "token" {
					raw["token"] = recordInfo.token
				} else {
					value := gjson.Get(string(resultJSON), f)
					//fmt.Printf("result %s -> %s\n", f, value)
					/*
						var raw2 map[string]interface{}
						err = json.Unmarshal([]byte(value.String()), &raw2)
						if err != nil {
							fmt.Printf("Err: %s\n", err)
						}
					*/
					raw[f] = value.Value()
				}
			}
			resultJSON, _ = json.Marshal(raw)
		}
	}
	//fmt.Fprintf(w, "<html><head><title>title</title></head>")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	var str string

	if len(recordInfo.appName) > 0 {
		str = fmt.Sprintf(`{"status":"ok","app":"%s","data":%s}`,
			recordInfo.appName, resultJSON)
	} else if len(recordInfo.session) > 0 {
		str = fmt.Sprintf(`{"status":"ok","session":"%s","data":%s}`,
			recordInfo.session, resultJSON)
	} else {
		str = fmt.Sprintf(`{"status":"ok","data":%s}`, resultJSON)
	}

	log.Printf("result: %s\n", str)
	w.Write([]byte(str))
}
