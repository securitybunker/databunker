package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	//"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) agreementAccept(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := audit("agreement accept for "+brief, identity, mode, identity)
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
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if exists == false {
		returnError(w, r, "not found", 404, nil, event)
		return
	}
	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, err := e.db.lookupUserRecord(identity)
		if err != nil || userBson == nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if e.enforceAuth(w, r, event) == "" {
			return
		}
		userTOKEN = identity
	} else {
		userBson, err := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		} else {
			if mode == "login" {
				returnError(w, r, "internal error", 405, nil, event)
				return
			}
			// else user not found - we allow to save consent for unlinked users!
		}
	}

	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	starttime := int32(0)
	expiration := int32(0)
	referencecode := getStringValue(records["referencecode"])
	lastmodifiedby := getStringValue(records["lastmodifiedby"])
	agreementmethod := getStringValue(records["agreementmethod"])
	status := getStringValue(records["status"])
	if len(status) == 0 {
		status = "yes"
	} else {
		status = normalizeConsentStatus(status)
	}
	if value, ok := records["expiration"]; ok {
		switch records["expiration"].(type) {
		case string:
			expiration, _ = parseExpiration(value.(string))
		case float64:
			expiration = int32(value.(float64))
		}
	}
	if value, ok := records["starttime"]; ok {
		switch records["starttime"].(type) {
		case string:
			starttime, _ = parseExpiration(value.(string))
		case float64:
			starttime = int32(value.(float64))
		}
	}
	switch mode {
	case "email":
		identity = normalizeEmail(identity)
	case "phone":
		identity = normalizePhone(identity, e.conf.Sms.DefaultCountry)
	}
	fmt.Printf("Processing agreement, status: %s\n", status)
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
	event := audit("consent withdraw for "+brief, identity, mode, identity)
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
	lbasis, err := e.db.getLegalBasis(brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if lbasis == nil {
		returnError(w, r, "not  found", 405, nil, event)
		return
	}
	userTOKEN := ""
	authResult := ""
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(identity)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		authResult = e.enforceAuth(w, r, event)
		if authResult == "" {
			return
		}
		userTOKEN = identity
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		} else {
			if mode == "login" {
				returnError(w, r, "internal error", 405, nil, event)
				return
			}
			// else user not found - we allow to save consent for unlinked users!
		}
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, event)
		return
	}
	lastmodifiedby := getStringValue(records["lastmodifiedby"])
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
	if selfService == false {
		// user can change consent only for briefs defined in self-service
		if len(authResult) == 0 {
			authResult = e.enforceAuth(w, r, event)
			if authResult == "" {
				return
			}
		}
	}

	if authResult == "login" && selfService == false {
		rtoken, rstatus, err := e.db.saveUserRequest("agreement-withdraw", userTOKEN, "", brief, nil, e.conf)
		if err != nil {
			returnError(w, r, "internal error", 405, err, event)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok","result":"%s","rtoken":"%s"}`, rstatus, rtoken)
		return
	}
	switch mode {
	case "email":
		identity = normalizeEmail(identity)
	case "phone":
		identity = normalizePhone(identity, e.conf.Sms.DefaultCountry)
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
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	brief = normalizeBrief(brief)
	if isValidBrief(brief) == false {
		returnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		returnError(w, r, "internal error", 405, nil, nil)
		return
	}
	if exists == false {
		returnError(w, r, "not found", 405, nil, nil)
		return
	}
	e.db.revokeLegalBasis(brief)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) getUserAgreements(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	mode := ps.ByName("mode")
	event := audit("privacy agreements for "+mode, identity, mode, identity)
	defer func() { event.submit(e.db) }()

	if validateMode(mode) == false {
		returnError(w, r, "bad mode", 405, nil, event)
		return
	}

	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(identity)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		if e.enforceAuth(w, r, event) == "" {
			return
		}
		userTOKEN = identity
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
			if e.enforceAuth(w, r, event) == "" {
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
	if e.enforceAuth(w, r, event) == "" {
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
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	//fmt.Printf("Total count of rows: %d\n", numRecords)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}

func (e mainEnv) getUserAgreement(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	identity := ps.ByName("identity")
	brief := ps.ByName("brief")
	mode := ps.ByName("mode")
	event := audit("privacy agreements for "+mode, identity, mode, identity)
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
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if exists == false {
		returnError(w, r, "not found", 404, nil, event)
	}
	userTOKEN := ""
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, _ := e.db.lookupUserRecord(identity)
		if userBson == nil {
			returnError(w, r, "internal error", 405, nil, event)
			return
		}
		if e.enforceAuth(w, r, event) == "" {
			return
		}
		userTOKEN = identity
	} else {
		// TODO: decode url in code!
		userBson, _ := e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
			if e.enforceAuth(w, r, event) == "" {
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
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	var resultJSON []byte
	resultJSON, err = e.db.viewAgreementRecord(userTOKEN, brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, err, event)
		return
	}
	//fmt.Printf("Total count of rows: %d\n", numRecords)
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
	event := audit("consent record for "+brief, identity, mode, identity)
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
	userTOKEN := identity
	var userBson bson.M
	if mode == "token" {
		if enforceUUID(w, identity, event) == false {
			return
		}
		userBson, _ = e.db.lookupUserRecord(identity)
	} else {
		userBson, _ = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
		if userBson != nil {
			userTOKEN = userBson["token"].(string)
			event.Record = userTOKEN
		}
	}
	if userBson == nil {
		returnError(w, r, "internal error", 405, nil, event)
		return
	}
	// make sure that user is logged in here, unless he wants to cancel emails
	if e.enforceAuth(w, r, event) == "" {
		return
	}
	resultJSON, err := e.db.viewConsentRecord(userTOKEN, brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	if resultJSON == nil {
		returnError(w, r, "not found", 405, nil, event)
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
	event := audit("consent get all for "+brief, brief, "brief", brief)
	defer func() { event.submit(e.db) }()
	if e.enforceAuth(w, r, event) == "" {
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}

*/
