package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
)

func (e mainEnv) userCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	event := CreateAuditEvent("create user record", "", "", "")
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if e.conf.Generic.CreateUserWithoutAccessToken == false {
		// anonymous user can not create user record, check token
		if e.EnforceAdmin(w, r, event) == "" {
			log.Println("Failed to create user, access denied, try to configure Create_user_without_access_token")
			return
		}
	}
	userJSON, err := utils.GetUserJSONStruct(r, e.conf.Sms.DefaultCountry)
	if err != nil {
		ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if len(userJSON.JsonData) == 0 {
		ReturnError(w, r, "empty request body", 405, nil, event)
		return
	}
	err = validateUserRecord(userJSON.JsonData)
	if err != nil {
		ReturnError(w, r, "user schema error: "+err.Error(), 405, err, event)
		return
	}
	// make sure that login, email and phone are unique
	if len(userJSON.LoginIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("login", userJSON.LoginIdx, e.conf)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			ReturnError(w, r, "duplicate index: login", 405, nil, event)
			return
		}
	}
	if len(userJSON.EmailIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("email", userJSON.EmailIdx, e.conf)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			ReturnError(w, r, "duplicate index: email", 405, nil, event)
			return
		}
	}
	if len(userJSON.PhoneIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("phone", userJSON.PhoneIdx, e.conf)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			ReturnError(w, r, "duplicate index: phone", 405, nil, event)
			return
		}
	}
	if len(userJSON.CustomIdx) > 0 {
		otherUserBson, err := e.db.lookupUserRecordByIndex("custom", userJSON.CustomIdx, e.conf)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		if otherUserBson != nil {
			ReturnError(w, r, "duplicate index: custom", 405, nil, event)
			return
		}
	}
	if len(userJSON.LoginIdx) == 0 &&
		len(userJSON.EmailIdx) == 0 &&
		len(userJSON.PhoneIdx) == 0 &&
		len(userJSON.CustomIdx) == 0 {
		ReturnError(w, r, "failed to create user, all user lookup fields are missing", 405, err, event)
		return
	}

	userTOKEN, err := e.db.createUserRecord(userJSON, event)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	encPhoneIdx := ""
	if len(userJSON.EmailIdx) > 0 {
		encEmailIdx, _ := utils.BasicStringEncrypt(userJSON.EmailIdx, e.db.masterKey, e.db.GetCode())
		e.db.linkAgreementRecords(userTOKEN, encEmailIdx)
	}
	if len(userJSON.PhoneIdx) > 0 {
		encPhoneIdx, _ = utils.BasicStringEncrypt(userJSON.PhoneIdx, e.db.masterKey, e.db.GetCode())
		e.db.linkAgreementRecords(userTOKEN, encPhoneIdx)
	}
	if len(userJSON.EmailIdx) > 0 && len(userJSON.PhoneIdx) > 0 {
		// delete duplicate consent records for user
		records, _ := e.db.store.GetList(storage.TblName.Agreements, "token", userTOKEN, 0, 0, "")
		var briefCodes []string
		for _, val := range records {
			if utils.SliceContains(briefCodes, val["brief"].(string)) == true {
				e.db.store.DeleteRecord2(storage.TblName.Agreements, "token", userTOKEN, "who", encPhoneIdx)
			} else {
				briefCodes = append(briefCodes, val["brief"].(string))
			}
		}
	}
	event.Record = userTOKEN
	utils.ReturnUUID(w, userTOKEN)
	notifyURL := e.conf.Notification.NotificationURL
	notifyProfileNew(notifyURL, userJSON.JsonData, "token", userTOKEN)
}

func (e mainEnv) userGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := CreateAuditEvent("get user record by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, userBSON, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	resultJSON, _ := e.db.userProfileDecrypt(userBSON)
	if resultJSON == nil {
		resultJSON = []byte("{}")
	}
	// if resultJSON == nil {
	// 	ReturnError(w, r, "record not found", 405, nil, event)
	// 	return
	// }
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
	event := CreateAuditEvent("change user record by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, userBSON, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	userJSON, _ := e.db.userProfileDecrypt(userBSON)
	if userJSON == nil {
		ReturnError(w, r, "user record not found", 405, nil, event)
		return
	}

	postData, err := utils.GetJSONPostData(r)
	if err != nil {
		ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	if postData == nil {
		ReturnError(w, r, "empty request body", 405, nil, event)
		return
	}

	authResult := e.EnforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	adminRecordChanged := false
	if UserSchemaEnabled() {
		adminRecordChanged, err = e.db.validateUserRecordChange(userJSON, postData, userTOKEN, authResult)
		if err != nil {
			ReturnError(w, r, "schema validation error: "+err.Error(), 405, err, event)
			return
		}
	}
	if authResult == "login" {
		event.Title = "user change-profile request"
		if e.conf.SelfService.UserRecordChange == false || adminRecordChanged == true {
			rtoken, rstatus, err := e.db.saveUserRequest("change-profile", userTOKEN, userBSON, "", "", postData, e.conf)
			if err != nil {
				ReturnError(w, r, "internal error", 405, err, event)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
			return
		}
	}
	oldJSON, newJSON, lookupErr, err := e.db.updateUserRecord(postData, userTOKEN, userBSON, event, e.conf)
	if lookupErr {
		ReturnError(w, r, "record not found", 405, errors.New("record not found"), event)
		return
	}
	if err != nil {
		ReturnError(w, r, "error updating user", 405, err, event)
		return
	}
	utils.ReturnUUID(w, userTOKEN)
	notifyURL := e.conf.Notification.NotificationURL
	notifyProfileChange(notifyURL, oldJSON, newJSON, "token", userTOKEN)
}

// user forgetme request comes here
func (e mainEnv) userDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := CreateAuditEvent("delete user record by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	strictCheck := true
	if mode == "email" {
		strictCheck = false
	}
	userTOKEN, userBSON, err := e.getUserToken(w, r, mode, identity, event, strictCheck)
	if strictCheck == true && userTOKEN == "" {
		return
	}
	if err != nil {
		return
	}
	authResult := e.EnforceAuth(w, r, event)
	if authResult == "" {
		return
	}
	if len(userTOKEN) == 0 {
		if authResult == "root" && mode == "email" {
			e.globalUserDelete(identity)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
			return
		}
		ReturnError(w, r, "record not found", 405, nil, event)
		return
	}

	if authResult == "login" {
		event.Title = "user forget-me request"
		if e.conf.SelfService.ForgetMe == false {
			rtoken, rstatus, err := e.db.saveUserRequest("forget-me", userTOKEN, userBSON, "", "", nil, e.conf)
			if err != nil {
				ReturnError(w, r, "internal error", 405, err, event)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
			return
		}
	}
	// decrypt user!
	userJSON, err := e.db.userProfileDecrypt(userBSON)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if userJSON == nil {
		userJSON = []byte("{}")
	} else {
		email := utils.GetStringValue(userBSON["email"])
		if len(email) > 0 {
			e.globalUserDelete(email)
		}
	}
	//fmt.Printf("deleting user %s\n", userTOKEN)
	_, err = e.db.deleteUserRecord(userJSON, userTOKEN, e.conf)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
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
	event := CreateAuditEvent("user prelogin by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	code0, err := decryptCaptcha(captcha)
	if err != nil || code0 != code {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"error","result":"captcha-error"}`)
		return
	}
	if mode != "phone" && mode != "email" {
		ReturnError(w, r, "bad mode", 405, nil, event)
		return
	}
	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if userBson != nil {
		userTOKEN := utils.GetUuidString(userBson["token"])
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
			//ReturnError(w, r, "record not found", 405, errors.New("record not found"), event)
			captcha, _ := generateCaptcha()
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(403)
			fmt.Fprintf(w, `{"status":"error","result":"record not found","captchaurl":"%s"}`, captcha)
			return
		}
		log.Println("User record not found, returning ok status")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","result":"done"}`)
}

func (e mainEnv) userLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmp := utils.Atoi(ps.ByName("tmp"))
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := CreateAuditEvent("user login by "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if mode != "phone" && mode != "email" {
		ReturnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
	if userBson == nil || err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}

	userTOKEN := utils.GetUuidString(userBson["token"])
	event.Record = userTOKEN
	tmpCode := int32(0)
	if _, ok := userBson["tempcode"]; ok {
		tmpCode = userBson["tempcode"].(int32)
	}
	if tmp == tmpCode {
		// user ented correct key
		// generate temp user access code
		xtoken, hashedToken, err := e.db.genUserLoginXtoken(userTOKEN)
		//fmt.Printf("generate user access token: %s\n", xtoken)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		event.Msg = "generated: " + hashedToken
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok","xtoken":"%s","token":"%s"}`, xtoken, userTOKEN)
		return
	}
	ReturnError(w, r, "internal error", 405, nil, event)
}
