package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) userNew(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	event := audit("create user record", "", "", "")
	defer func() { event.submit(e.db) }()

	if e.conf.Generic.Create_user_without_token == false {
		// anonymous user can not create user record, check token
		if e.enforceAuth(w, r, event) == false {
			fmt.Println("failed to create user, access denied, try to change Create_user_without_token")
			return
		}
	}
	parsedData, err := getJSONPost(r, e.conf.Sms.Default_country)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(parsedData.jsonData) == 0 {
		returnError(w, r, "empty body", 405, nil, event)
		return
	}
	// make sure that login, email and phone are unique
	if len(parsedData.loginIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("login", parsedData.loginIdx, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			returnError(w, r, "duplicate index: login", 405, nil, event)
			return
		}
	}
	if len(parsedData.emailIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("email", parsedData.emailIdx, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			returnError(w, r, "duplicate index: email", 405, nil, event)
			return
		}
	}
	if len(parsedData.phoneIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("phone", parsedData.phoneIdx, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			returnError(w, r, "duplicate index: phone", 405, nil, event)
			return
		}
	}
	userTOKEN, err := e.db.createUserRecord(parsedData, event)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	event.Record = userTOKEN
	returnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var resultJSON []byte
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("get user record by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()
	if e.enforceAuth(w, r, event) == false {
		return
	}
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := address
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		resultJSON, err = e.db.getUser(address)
	} else {
		resultJSON, userTOKEN, err = e.db.getUserIndex(address, mode, e.conf)
		event.Record = userTOKEN
	}
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "record not found", 405, nil, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s","data":%s}`, userTOKEN, resultJSON)
	fmt.Printf("record: %s\n", finalJSON)
	//fmt.Fprintf(w, "<html><head><title>title</title></head>")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

func (e mainEnv) userChange(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("change user record by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if e.enforceAuth(w, r, event) == false {
		return
	}
	if validateMode(mode) == false {
		returnError(w, r, "bad index", 405, nil, event)
		return
	}
	if mode == "token" && enforceUUID(w, address, event) == false {
		return
	}
	parsedData, err := getJSONPost(r, e.conf.Sms.Default_country)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(parsedData.jsonData) == 0 {
		returnError(w, r, "empty body", 405, nil, event)
		return
	}
	userTOKEN := address
	if mode != "token" {
		userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		userTOKEN = userBson["token"].(string)
		event.Record = userTOKEN
	}
	err = e.db.updateUserRecord(parsedData, userTOKEN, event, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	return
}

func (e mainEnv) userDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("delete user record by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if e.enforceAuth(w, r, event) == false {
		return
	}
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	if mode == "token" && enforceUUID(w, address, event) == false {
		return
	}
	userTOKEN := address
	if mode != "token" {
		userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		userTOKEN = userBson["token"].(string)
		event.Record = userTOKEN
	}
	fmt.Printf("deleting user %s", userTOKEN)
	result, err := e.db.deleteUserRecord(userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if result == false {
		// user deleted
		event.Status = "failed"
		event.Msg = "failed to delete"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}

func (e mainEnv) userLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("user login by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if mode != "phone" && mode != "email" {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if userBson != nil {
		userTOKEN := userBson["token"].(string)
		event.Record = userTOKEN
		if address == "4444" || address == "test@paranoidguy.com" {
			// check if it is demo account.
			// the address is always 4444
			// no need to send any notifications
			e.db.generateDemoLoginCode(userTOKEN)
		} else {
			rnd := e.db.generateTempLoginCode(userTOKEN)
			if mode == "email" {
				go sendCodeByEmail(rnd, address, e.conf)
			} else if mode == "phone" {
				go sendCodeByPhone(rnd, address, e.conf)
			}
		}
	} else {
		fmt.Println("user record not found, stil returning ok status")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}

func (e mainEnv) userLoginEnter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmp := ps.ByName("tmp")
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("user login by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if mode != "phone" && mode != "email" {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}

	if userBson != nil {
		userTOKEN := userBson["token"].(string)
		event.Record = userTOKEN
		fmt.Printf("Found user record: %s\n", userTOKEN)
		tmpCode := userBson["tempcode"].(string)
		if tmp == tmpCode {
			// user ented correct key
			// generate temp user access code
			xtoken, err := e.db.generateUserLoginXToken(userTOKEN)
			fmt.Printf("generate user access token: %s\n", xtoken)
			if err != nil {
				returnError(w, r, "internal error", 405, err, event)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","xtoken":"%s","token":"%s"}`, xtoken, userTOKEN)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":""}`)
}
