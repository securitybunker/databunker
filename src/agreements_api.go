package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/utils"
	//"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) agreementAccept(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")

	brief = utils.NormalizeBrief(brief)
	event := CreateAuditEvent("accept agreement for "+brief, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	strictCheck := true
	if mode == "phone" || mode == "email" {
		strictCheck = false
	}
	userTOKEN, _, err := e.getUserToken(w, r, mode, identity, event, strictCheck)
	if strictCheck == true && userTOKEN == "" {
		return
	}
	if err != nil {
		return
	}

	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "incorrect brief format", 405, nil, event)
		return
	}
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if exists == false {
		ReturnError(w, r, "not found", 404, nil, event)
		return
	}
	postData, err := utils.GetJSONPostMap(r)
	if err != nil {
		ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	starttime := int32(0)
	expiration := int32(0)
	referencecode := utils.GetStringValue(postData["referencecode"])
	lastmodifiedby := utils.GetStringValue(postData["lastmodifiedby"])
	agreementmethod := utils.GetStringValue(postData["agreementmethod"])
	status := utils.GetStringValue(postData["status"])
	if len(status) == 0 {
		status = "yes"
	} else {
		status = utils.NormalizeConsentValue(status)
	}
	if value, ok := postData["expiration"]; ok {
		switch postData["expiration"].(type) {
		case string:
			expiration, _ = utils.ParseExpiration(value.(string))
		case float64:
			expiration = int32(value.(float64))
		}
	}
	if value, ok := postData["starttime"]; ok {
		switch postData["starttime"].(type) {
		case string:
			starttime, _ = utils.ParseExpiration(value.(string))
		case float64:
			starttime = int32(value.(float64))
		}
	}
	switch mode {
	case "email":
		identity = utils.NormalizeEmail(identity)
	case "phone":
		identity = utils.NormalizePhone(identity, e.conf.Sms.DefaultCountry)
	}

	log.Printf("Processing agreement, status: %s\n", status)
	e.db.acceptAgreement(userTOKEN, mode, identity, brief, status, agreementmethod,
		referencecode, lastmodifiedby, starttime, expiration)
	/*
		notifyURL := e.conf.Notification.NotificationURL
		if newStatus == true && len(notifyURL) > 0 {
			// change notificate on new record or if status change
			if len(userTOKEN) > 0 {
				notifyConsentChange(notifyURL, brief, status, "token", userTOKEN)
			} else {
				notifyConsentChange(notifyURL, brief, status, mode, identity)
			}
		}
	*/
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) agreementWithdraw(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")

	brief = utils.NormalizeBrief(brief)
	event := CreateAuditEvent("withdraw agreement for "+brief, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	strictCheck := true
	if mode == "phone" || mode == "email" {
		strictCheck = false
	}
	userTOKEN, userBSON, err := e.getUserToken(w, r, mode, identity, event, strictCheck)
	if strictCheck == true && userTOKEN == "" {
		//log.Printf("Strict check err")
		return
	}
	if err != nil {
		return
	}
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, event)
		return
	}
	lbasis, err := e.db.getLegalBasis(brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if lbasis == nil {
		ReturnError(w, r, "not  found", 405, nil, event)
		return
	}
	postData, err := utils.GetJSONPostMap(r)
	if err != nil {
		ReturnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	lastmodifiedby := utils.GetStringValue(postData["lastmodifiedby"])
	selfService := false
	if value, ok := lbasis["usercontrol"]; ok {
		if reflect.TypeOf(value).Kind() == reflect.Bool {
			selfService = value.(bool)
		} else {
			num := value.(int32)
			if num > 0 {
				selfService = true
			}
		}
	}
	authResult := e.EnforceAuth(w, r, event)
	if selfService == false {
		// user can change consent only for briefs defined in self-service
		if len(authResult) == 0 {
			if e.EnforceAdmin(w, r, event) == "" {
				return
			}
		}
	}
	if authResult == "login" && selfService == false {
		rtoken, rstatus, err := e.db.saveUserRequest("agreement-withdraw", userTOKEN, userBSON, "", brief, nil, e.conf)
		if err != nil {
			ReturnError(w, r, "internal error", 405, err, event)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
		return
	}
	switch mode {
	case "email":
		identity = utils.NormalizeEmail(identity)
	case "phone":
		identity = utils.NormalizePhone(identity, e.conf.Sms.DefaultCountry)
	}
	e.db.withdrawAgreement(userTOKEN, brief, mode, identity, lastmodifiedby)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
	notifyURL := e.conf.Notification.NotificationURL
	if len(userTOKEN) > 0 {
		notifyConsentChange(notifyURL, brief, "no", "token", userTOKEN)
	} else {
		notifyConsentChange(notifyURL, brief, "no", mode, identity)
	}
}

func (e mainEnv) agreementRevokeAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	brief := ps.ByName("brief")
	brief = utils.NormalizeBrief(brief)
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, nil, nil)
		return
	}
	if exists == false {
		ReturnError(w, r, "not found", 405, nil, nil)
		return
	}
	e.db.revokeLegalBasis(brief)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) agreementList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := CreateAuditEvent("get agreements for "+mode, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, _, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	var resultJSON []byte
	var numRecords int
	var err error
	if len(userTOKEN) > 0 {
		resultJSON, numRecords, err = e.db.listAgreementRecords(userTOKEN)
	} else {
		resultJSON, numRecords, err = e.db.listAgreementRecordsByIdentity(identity)
	}

	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) agreementGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")

	brief = utils.NormalizeBrief(brief)
	event := CreateAuditEvent("get agreements for "+brief, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	userTOKEN, _, _ := e.getUserToken(w, r, mode, identity, event, true)
	if userTOKEN == "" {
		return
	}
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, event)
		return
	}
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if exists == false {
		ReturnError(w, r, "not found", 404, nil, event)
		return
	}

	resultJSON, err := e.db.viewAgreementRecord(userTOKEN, brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		ReturnError(w, r, "not found", 405, err, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","data":%s}`, resultJSON)
	w.Write([]byte(str))
}

/*
func (e mainEnv) consentUserRecord(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := CreateAuditEvent("consent record for "+brief, identity, mode, identity)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()

	if utils.ValidateMode(mode) == false {
		ReturnError(w, r, "bad mode", 405, nil, event)
		return
	}
	brief = utils.NormalizeBrief(brief)
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, event)
		return
	}
	userTOKEN := identity
	var userBson bson.M
	if mode == "token" {
		if EnforceUUID(w, identity, event) == false {
			return
		}
		userBson, _ = e.db.lookupUserRecord(identity)
	} else {
		userBson, _ = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = utils.GetUuidString(userBson["token"])
			event.Record = userTOKEN
		}
	}
	if userBson == nil {
		ReturnError(w, r, "internal error", 405, nil, event)
		return
	}
	// make sure that user is logged in here, unless he wants to cancel emails
	if e.EnforceAuth(w, r, event) == "" {
		return
	}
	resultJSON, err := e.db.viewConsentRecord(userTOKEN, brief)
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		ReturnError(w, r, "not found", 405, nil, event)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","data":%s}`, resultJSON)
	w.Write([]byte(str))
}
*/

/*
func (e mainEnv) consentFilterRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	brief := ps.ByName("brief")
	event := CreateAuditEvent("consent get all for "+brief, brief, "brief", brief)
	defer func() { SaveAuditEvent(event, e.db, e.conf) }()
	if e.EnforceAuth(w, r, event) == "" {
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
		ReturnError(w, r, "internal error", 405, err, event)
		return
	}
	log.Printf("Total count of rows: %d\n", numRecords)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}

*/
