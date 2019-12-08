package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
)

func (e mainEnv) userNewToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit("create user temp access xtoken", userTOKEN)
	defer func() { event.submit(e.db) }()

	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
		return
	}
	records, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	fields := ""
	expiration := ""
	appName := ""
	if value, ok := records["fields"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			fields = value.(string)
		}
	}
	if value, ok := records["expiration"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			expiration = value.(string)
		} else {
			returnError(w, r, "failed to parse expiration field", 405, err, event)
			return
		}
	}
	if value, ok := records["app"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			appName = strings.ToLower(value.(string))
			if len(appName) > 0 && isValidApp(appName) == false {
				returnError(w, r, "failed to parse app field", 405, nil, event)
			}
		} else {
			// type is different
			returnError(w, r, "failed to parse app field", 405, nil, event)
		}
	}
	if len(expiration) == 0 {
		returnError(w, r, "missing expiration field", 405, err, event)
		return
	}
	xtokenUUID, err := e.db.generateUserTempXToken(userTOKEN, fields, expiration, appName)
	if err != nil {
		fmt.Println(err)
		returnError(w, r, err.Error(), 405, err, event)
		return
	}
	event.Msg = "Generated " + xtokenUUID
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","xtoken":%q}`, xtokenUUID)
}

func (e mainEnv) userCheckToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	event := audit("get record by user temp access token", "")
	defer func() { event.submit(e.db) }()

	xtoken := ps.ByName("xtoken")
	if enforceUUID(w, xtoken, event) == false {
		return
	}
	authResult, err := e.db.checkUserAuthXToken(xtoken)
	if err != nil {
		fmt.Printf("%d access denied for : %s\n", http.StatusForbidden, xtoken)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}
	var resultJSON []byte
	if len(authResult.token) > 0 {
		event.Record = authResult.token
		event.App = authResult.appName
		fmt.Printf("displaying fields: %s, user token: %s\n", authResult.fields, authResult.token)

		if len(authResult.appName) > 0 {
			resultJSON, err = e.db.getUserApp(authResult.token, authResult.appName)
		} else {
			resultJSON, err = e.db.getUser(authResult.token)
		}
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if resultJSON == nil {
			returnError(w, r, "not found", 405, err, event)
			return
		}
		fmt.Printf("Full user json: %s\n", resultJSON)
		if len(authResult.fields) > 0 {
			raw := make(map[string]interface{})
			//var newJSON json
			allFields := parseFields(authResult.fields)
			for _, f := range allFields {
				if f == "token" {
					raw["token"] = authResult.token
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
	if len(resultJSON) == 0 {
		str = fmt.Sprintf(`{"status":"ok","type":"%s"}`, authResult.ttype)
	} else {
		if len(authResult.appName) > 0 {
			str = fmt.Sprintf(`{"status":"ok","type":"%s","app":"%s","data":%s}`,
				authResult.ttype, authResult.appName, resultJSON)
		} else {
			str = fmt.Sprintf(`{"status":"ok","type":"%s","data":%s}`,
				authResult.ttype, resultJSON)
		}
	}
	fmt.Printf("result: %s\n", str)
	w.Write([]byte(str))
}
