package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/paranoidguy/databunker/src/storage"
)

func (e mainEnv) userNew(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	event := audit("create user record", "", "", "")
	defer func() { event.submit(e.db) }()

	if e.conf.Generic.CreateUserWithoutAccessToken == false {
		// anonymous user can not create user record, check token
		if e.enforceAuth(w, r, event) == "" {
			fmt.Println("failed to create user, access denied, try to change Create_user_without_access_token")
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
	}
	parsedData, err := getJSONPost(r, e.conf.Sms.DefaultCountry)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(parsedData.jsonData) == 0 {
		returnError(w, r, "empty request body", 405, nil, event)
		return
	}
	err = validateUserRecord(parsedData.jsonData)
	if err != nil {
		returnError(w, r, "user schema error: "+err.Error(), 405, err, event)
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
	if len(parsedData.emailIdx) > 0 {
		e.db.linkAgreementRecords(userTOKEN, "email", parsedData.emailIdx)
	}
	if len(parsedData.phoneIdx) > 0 {
		e.db.linkAgreementRecords(userTOKEN, "phone", parsedData.phoneIdx)
	}
	if len(parsedData.emailIdx) > 0 && len(parsedData.phoneIdx) > 0 {
		// delete duplicate consent records for user
		records, _ := e.db.store.GetList(storage.TblName.Agreements, "who", parsedData.emailIdx, 0, 0, "")
		var briefCodes []string
		for _, val := range records {
			//fmt.Printf("adding brief code: %s\n", val["brief"].(string))
			briefCodes = append(briefCodes, val["brief"].(string))
		}
		records, _ = e.db.store.GetList(storage.TblName.Agreements, "who", parsedData.phoneIdx, 0, 0, "")
		for _, val := range records {
			//fmt.Printf("XXX checking brief code for duplicates: %s\n", val["brief"].(string))
			if contains(briefCodes, val["brief"].(string)) == true {
				e.db.store.DeleteRecord2(storage.TblName.Agreements, "token", userTOKEN, "who", parsedData.phoneIdx)
			}
		}
	}
	event.Record = userTOKEN
	returnUUID(w, userTOKEN)
	notifyURL := e.conf.Notification.NotificationURL
	notifyProfileNew(notifyURL, parsedData.jsonData, "token", userTOKEN)
	return
}

func (e mainEnv) userGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var resultJSON []byte
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("get user record by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := ""
	authResult := ""
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		resultJSON, err = e.db.getUser(address)
		userTOKEN = address
	} else {
		resultJSON, userTOKEN, err = e.db.getUserIndex(address, mode, e.conf)
		event.Record = userTOKEN
	}
	if err != nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	authResult = e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if resultJSON == nil {
		returnError(w, r, "record not found", 405, nil, event)
		return
	}
	finalJSON := fmt.Sprintf(`{"status":"ok","token":"%s","data":%s}`, userTOKEN, resultJSON)
	//fmt.Printf("record: %s\n", finalJSON)
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

	if validateMode(mode) == false {
		returnError(w, r, "bad index", 405, nil, event)
		return
	}

	parsedData, err := getJSONPost(r, e.conf.Sms.DefaultCountry)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(parsedData.jsonData) == 0 {
		returnError(w, r, "empty request body", 405, nil, event)
		return
	}

	userTOKEN := ""
	var userJSON []byte
	if mode == "token" {
		if enforceUUID(w, address, event) == false {
			return
		}
		userTOKEN = address
		userJSON, err = e.db.getUser(address)
	} else {
		userJSON, userTOKEN, err = e.db.getUserIndex(address, mode, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if userJSON == nil {
			returnError(w, r, "record not found", 405, nil, event)
			return
		}
		event.Record = userTOKEN
	}
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	adminRecordChanged := false
	if UserSchemaEnabled() {
	  adminRecordChanged, err = e.db.validateUserRecordChange(userJSON, parsedData.jsonData, userTOKEN, authResult)
	  if err != nil {
	    returnError(w, r, "schema validation error: " + err.Error(), 405, err, event)
		return
	  }
	}
	if authResult == "login" {
		event.Title = "user change-profile request"
		if e.conf.SelfService.UserRecordChange == false || adminRecordChanged == true {
			rtoken, rstatus, err := e.db.saveUserRequest("change-profile", userTOKEN, "", "", parsedData.jsonData)
			if err != nil {
				returnError(w, r, "internal error", 405, err, event)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
			return
		}
	}
	oldJSON, newJSON, lookupErr, err := e.db.updateUserRecord(parsedData.jsonData, userTOKEN, event, e.conf)
	if lookupErr {
		returnError(w, r, "record not found", 405, errors.New("record not found"), event)
		return
	}
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	notifyURL := e.conf.Notification.NotificationURL
	notifyProfileChange(notifyURL, oldJSON, newJSON, "token", userTOKEN)
}

// user forgetme request comes here
func (e mainEnv) userDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("delete user record by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	var err error
	var resultJSON []byte
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
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if resultJSON == nil {
		returnError(w, r, "record not found", 405, nil, event)
		return
	}

	if authResult == "login" {
		event.Title = "user forget-me request"
		if e.conf.SelfService.ForgetMe == false {
			rtoken, rstatus, err := e.db.saveUserRequest("forget-me", userTOKEN, "", "", nil)
			if err != nil {
				returnError(w, r, "internal error", 405, err, event)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
			return
		}
	}
	e.globalUserDelete(userTOKEN)
	//fmt.Printf("deleting user %s\n", userTOKEN)
	_, err = e.db.deleteUserRecord(resultJSON, userTOKEN)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
	notifyURL := e.conf.Notification.NotificationURL
	notifyForgetMe(notifyURL, resultJSON, "token", userTOKEN)
}

func (e mainEnv) userPrelogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("user prelogin by "+mode, address, mode, address)
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
		if mode == "email" {
			//notifyURL := e.conf.Notification.NotificationURL
			//notifyBadLogin(notifyURL, mode, address)
			e.pluginUserLookup(address)
			returnError(w, r, "record not found", 405, errors.New("record not found"), event)
			return
		}
		fmt.Println("user record not found, still returning ok status")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}

func (e mainEnv) userLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmp := atoi(ps.ByName("tmp"))
	address := ps.ByName("address")
	mode := ps.ByName("mode")
	event := audit("user login by "+mode, address, mode, address)
	defer func() { event.submit(e.db) }()

	if mode != "phone" && mode != "email" {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userBson, err := e.db.lookupUserRecordByIndex(mode, address, e.conf)
	if userBson == nil || err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}

	userTOKEN := userBson["token"].(string)
	event.Record = userTOKEN
	tmpCode := int32(0)
	if _, ok := userBson["tempcode"]; ok {
		tmpCode = userBson["tempcode"].(int32)
	}
	if tmp == tmpCode {
		// user ented correct key
		// generate temp user access code
		xtoken, hashedToken, err := e.db.generateUserLoginXtoken(userTOKEN)
		//fmt.Printf("generate user access token: %s\n", xtoken)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		event.Msg = "generated: " + hashedToken
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok","xtoken":"%s","token":"%s"}`, xtoken, userTOKEN)
		return
	}
	returnError(w, r, "internal error", 405, nil, event)
}
