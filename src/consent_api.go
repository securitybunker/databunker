package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) consentAccept(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	code := ps.ByName("code")
	index := ps.ByName("index")
	event := audit("consent accept by "+index, code)
	defer func() { event.submit(e.db) }()

	userTOKEN := ""
	if index == "token" {
		if enforceUUID(w, code, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(code)
		if userBson != nil {
			userTOKEN = code
		}
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(index, code)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
		}
	}
	defer func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}()
	records, err := getJSONPostData(r)
	if err != nil {
		//returnError(w, r, "internal error", 405, err, event)
		return
	}
	brief := ""
	message := ""
	status := "accept"
	if value, ok := records["brief"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			brief = value.(string)
		}
	}
	if value, ok := records["message"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			message = value.(string)
		}
	}
	if value, ok := records["status"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			status = value.(string)
		}
	}
	if len(brief) == 0 {
		//returnError(w, r, "internal error", 405, nil, event)
		return
	}
	if len(message) == 0 {
		message = brief
	}
	e.db.createConsentRecord(userTOKEN, index, code, brief, message, status)
}

func (e mainEnv) consentCancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := ps.ByName("code")
	index := ps.ByName("index")
	event := audit("consent cancel by "+index, code)
	defer func() { event.submit(e.db) }()
	userTOKEN := code
	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	// make sure that user is logged in here, unless he wants to cancel emails
	if e.enforceAuth(w, r, event) == false {
		return
	}
	records, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	brief := ""
	if value, ok := records["brief"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			brief = value.(string)
		}
	}
	if len(brief) == 0 {
		returnError(w, r, "consent brief code is missing", 405, nil, event)
		return
	}
	e.db.cancelConsentRecord(userTOKEN, brief)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) consentList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := ps.ByName("code")
	index := ps.ByName("index")
	event := audit("consent list of events by "+index, code)
	defer func() { event.submit(e.db) }()
	userTOKEN := code
	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	// make sure that user is logged in here, unless he wants to cancel emails
	if e.enforceAuth(w, r, event) == false {
		return
	}

	resultJSON, numRecords, err := e.db.listConsentRecords(userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	fmt.Printf("Total count of rows: %d\n", numRecords)
	//fmt.Fprintf(w, "<html><head><title>title</title></head>")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}
