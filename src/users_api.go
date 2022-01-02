package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) userCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	if len(parsedData.customIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("custom", parsedData.customIdx, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			returnError(w, r, "duplicate index: custom", 405, nil, event)
			return
		}
	}
	if len(parsedData.loginIdx) == 0 &&
	   len(parsedData.emailIdx) == 0 &&
	   len(parsedData.phoneIdx) == 0 &&
	   len(parsedData.customIdx) == 0 {
		returnError(w, r, "failed to create user, all user lookup fields are missing", 405, err, event)
		return
	}

	userTOKEN, err := e.db.createUserRecord(parsedData, event)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	encPhoneIdx := ""
	if len(parsedData.emailIdx) > 0 {
		encEmailIdx, _ := basicStringEncrypt(parsedData.emailIdx, e.db.masterKey, e.db.GetCode())
		e.db.linkAgreementRecords(userTOKEN, encEmailIdx)
	}
	if len(parsedData.phoneIdx) > 0 {
		encPhoneIdx, _ = basicStringEncrypt(parsedData.phoneIdx, e.db.masterKey, e.db.GetCode())
		e.db.linkAgreementRecords(userTOKEN, encPhoneIdx)
	}
	if len(parsedData.emailIdx) > 0 && len(parsedData.phoneIdx) > 0 {
		// delete duplicate consent records for user
		records, _ := e.db.store.GetList(storage.TblName.Agreements, "token", userTOKEN, 0, 0, "")
		var briefCodes []string
		for _, val := range records {
			if contains(briefCodes, val["brief"].(string)) == true {
				e.db.store.DeleteRecord2(storage.TblName.Agreements, "token", userTOKEN, "who", encPhoneIdx)
			} else {
				briefCodes = append(briefCodes, val["brief"].(string))
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
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("get user record by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()
	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userTOKEN := ""
	authResult := ""
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		resultJSON, err = e.db.getUserJSON(identity)
		userTOKEN = identity
	} else {
		resultJSON, userTOKEN, err = e.db.getUserJSONByIndex(identity, mode, e.conf)
		event.Record = userTOKEN
	}
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
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
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("change user record by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad index", 405, nil, event)
		return
	}

	jsonData, err := getJSONPostData(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if jsonData == nil {
		returnError(w, r, "empty request body", 405, nil, event)
		return
	}
	userTOKEN := ""
	var userJSON []byte
	var userBSON bson.M
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userTOKEN = identity
		userJSON, userBSON, err = e.db.getUser(identity)
	} else {
		userJSON, userTOKEN, userBSON, err = e.db.getUserByIndex(identity, mode, e.conf)
		event.Record = userTOKEN
	}
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if userJSON == nil {
		returnError(w, r, "user record not found", 405, nil, event)
		return
	}
	authResult := e.enforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	adminRecordChanged := false
	if UserSchemaEnabled() {
		adminRecordChanged, err = e.db.validateUserRecordChange(userJSON, jsonData, userTOKEN, authResult)
		if err != nil {
			returnError(w, r, "schema validation error: "+err.Error(), 405, err, event)
			return
		}
	}
	if authResult == "login" {
		event.Title = "user change-profile request"
		if e.conf.SelfService.UserRecordChange == false || adminRecordChanged == true {
			rtoken, rstatus, err := e.db.saveUserRequest("change-profile", userTOKEN, "", "", jsonData, e.conf)
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
	oldJSON, newJSON, lookupErr, err := e.db.updateUserRecord(jsonData, userTOKEN, userBSON, event, e.conf)
	if lookupErr {
		returnError(w, r, "record not found", 405, errors.New("record not found"), event)
		return
	}
	if err != nil {
		returnError(w, r, "error updating user", 405, err, event)
		return
	}
	returnUUID(w, userTOKEN)
	notifyURL := e.conf.Notification.NotificationURL
	notifyProfileChange(notifyURL, oldJSON, newJSON, "token", userTOKEN)
}

// user forgetme request comes here
func (e mainEnv) userDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("delete user record by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	var err error
	var userBSON bson.M
	var userJSON []byte
	userTOKEN := identity
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userJSON, userBSON, err = e.db.getUser(identity)
	} else {
		userJSON, userTOKEN, userBSON, err = e.db.getUserByIndex(identity, mode, e.conf)
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
	if userJSON == nil {
		if authResult == "root" && mode == "email" {
			e.globalUserDelete(identity)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
		}
		returnError(w, r, "record not found", 405, nil, event)
		return
	}

	if authResult == "login" {
		event.Title = "user forget-me request"
		if e.conf.SelfService.ForgetMe == false {
			rtoken, rstatus, err := e.db.saveUserRequest("forget-me", userTOKEN, "", "", nil, e.conf)
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
	email := getStringValue(userBSON["email"])
	if len(email) > 0 {
		e.globalUserDelete(email)
	}
	//fmt.Printf("deleting user %s\n", userTOKEN)
	_, err = e.db.deleteUserRecord(userJSON, userTOKEN, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
	notifyURL := e.conf.Notification.NotificationURL
	notifyForgetMe(notifyURL, userJSON, "token", userTOKEN)
}

func (e mainEnv) userPrelogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	captcha := ps.ByName("captcha")
	code := ps.ByName("code")
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("user prelogin by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()

	code0, err := decryptCaptcha(captcha)
	if err != nil || code0 != code {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"error","result":"captcha-error"}`)
		return
	}
	if mode != "phone" && mode != "email" {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if userBson != nil {
		userTOKEN := userBson["token"].(string)
		event.Record = userTOKEN
		if identity == "4444" || identity == "test@securitybunker.io" {
			// check if it is demo account.
			// no need to send any notifications
			e.db.generateDemoLoginCode(userTOKEN)
		} else {
			rnd := e.db.generateTempLoginCode(userTOKEN)
			if mode == "email" {
				go sendCodeByEmail(rnd, identity, e.conf)
			} else if mode == "phone" {
				go sendCodeByPhone(rnd, identity, e.conf)
			}
		}
	} else {
		if mode == "email" {
			//notifyURL := e.conf.Notification.NotificationURL
			//notifyBadLogin(notifyURL, mode, identity)
			e.pluginUserLookup(identity)
			//returnError(w, r, "record not found", 405, errors.New("record not found"), event)
			captcha, _ := generateCaptcha()
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(403)
			fmt.Fprintf(w, `{"status":"error","result":"record not found","captchaurl":"%s"}`, captcha)
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
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("user login by "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()

	if mode != "phone" && mode != "email" {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
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
