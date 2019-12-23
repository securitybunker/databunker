package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) consentAccept(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := audit("consent accept for "+brief, address, mode, address)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	brief = normalizeBrief(brief)
	if isValidBrief(brief) == false {
		returnError(w, r, "bad brief format", 405, nil, event)
		return
	}

	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(address)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		if e.enforceAuth(w, r, event) == false {
			return
		}
		userTOKEN = address
	} else {
		userBson, _ := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
			if e.enforceAuth(w, r, event) == false {
				return
			}
		} else {
			if mode == "login" {
				returnError(w, r, "internal error", 405, nil, event)
				return
			}
			// else user not found - we allow to save consent for unlinked users!
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
	message := ""
	status := "accept"
	expiration := int32(0)
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
	if value, ok := records["expiration"]; ok {
		switch records["expiration"].(type) {
		case string:
			expiration = atoi(value.(string))
		case int:
			expiration = value.(int32)
		case int32:
			expiration = value.(int32)
		case int64:
			expiration = value.(int32)
		}
	}
	switch mode {
	case "email":
		address = normalizeEmail(address)
	case "phone":
		address = normalizePhone(address, e.conf.Sms.Default_country)
	}
	e.db.createConsentRecord(userTOKEN, mode, address, brief, message, status, expiration)
}

func (e mainEnv) consentCancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := audit("consent withdraw for "+brief, address, mode, address)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	brief = normalizeBrief(brief)
	if isValidBrief(brief) == false {
		returnError(w, r, "bad brief format", 405, nil, event)
		return
	}

	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(address)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		if e.enforceAuth(w, r, event) == false {
			return
		}
		userTOKEN = address
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
			if e.enforceAuth(w, r, event) == false {
				return
			}
		} else {
			if mode == "login" {
				returnError(w, r, "internal error", 405, nil, event)
				return
			}
			// else user not found - we allow to save consent for unlinked users!
		}
	}
	// make sure that user is logged in here, unless he wants to cancel emails
	//if e.enforceAuth(w, r, event) == false {
	//	return
	//}
	switch mode {
	case "email":
		address = normalizeEmail(address)
	case "phone":
		address = normalizePhone(address, e.conf.Sms.Default_country)
	}
	e.db.cancelConsentRecord(userTOKEN, brief, mode, address)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) consentAllUserRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("consent list of records for "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(address)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		if e.enforceAuth(w, r, event) == false {
			return
		}
		userTOKEN = address
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
			if e.enforceAuth(w, r, event) == false {
				return
			}
		} else {
			if mode == "login" {
				returnError(w, r, "internal error", 405, nil, event)
				return
			}
			// else user not found - we allow to save consent for unlinked users!

		}
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

func (e mainEnv) consentUserRecord(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := audit("consent record for "+brief, address, mode, address)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	brief = normalizeBrief(brief)
	if isValidBrief(brief) == false {
		returnError(w, r, "bad brief format", 405, nil, event)
		return
	}
	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(address)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		userTOKEN = address
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}

	// make sure that user is logged in here, unless he wants to cancel emails
	if e.enforceAuth(w, r, event) == false {
		return
	}

	resultJSON, err := e.db.viewConsentRecord(userTOKEN, brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","data":%s}`, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) consentFilterRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	brief := ps.ByName("brief")
	event := audit("consent get all for "+brief, brief, "brief", brief)
	defer func() { event.submit(e.db) }()
	if e.enforceAuth(w, r, event) == false {
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
	resultJSON, numRecords, err := e.db.filterConsentRecords(brief, offset, limit)
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
