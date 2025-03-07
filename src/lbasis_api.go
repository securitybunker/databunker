package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/utils"
	//"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) legalBasisCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	brief := ps.ByName("brief")
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	brief = utils.NormalizeBrief(brief)
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	postData, err := utils.GetJSONPostMap(r)
	if err != nil {
		ReturnError(w, r, "failed to decode request body", 405, err, nil)
		return
	}
	newbrief := utils.GetStringValue(postData["brief"])
	if len(newbrief) > 0 && newbrief != brief {
		if utils.CheckValidBrief(newbrief) == false {
			ReturnError(w, r, "bad brief format", 405, nil, nil)
			return
		}
	}
	status := utils.GetStringValue(postData["status"])
	module := utils.GetStringValue(postData["module"])
	fulldesc := utils.GetStringValue(postData["fulldesc"])
	shortdesc := utils.GetStringValue(postData["shortdesc"])
	basistype := utils.GetStringValue(postData["basistype"])
	requiredmsg := utils.GetStringValue(postData["requiredmsg"])
	usercontrol := false
	requiredflag := false
	if status != "disabled" {
		status = "active"
	}
	if value, ok := postData["usercontrol"]; ok {
		if reflect.TypeOf(value).Kind() == reflect.Bool {
			usercontrol = value.(bool)
		} else {
			num := value.(int32)
			if num > 0 {
				usercontrol = true
			}
		}
	}
	if value, ok := postData["requiredflag"]; ok {
		if reflect.TypeOf(value).Kind() == reflect.Bool {
			requiredflag = value.(bool)
		} else {
			num := value.(int32)
			if num > 0 {
				requiredflag = true
			}
		}
	}
	e.db.createLegalBasis(brief, newbrief, module, shortdesc, fulldesc, basistype, requiredmsg, status, usercontrol, requiredflag)
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

func (e mainEnv) legalBasisDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	brief := ps.ByName("brief")
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	brief = utils.NormalizeBrief(brief)
	if utils.CheckValidBrief(brief) == false {
		ReturnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	e.db.unlinkProcessingActivityBrief(brief)
	e.db.deleteLegalBasis(brief)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) legalBasisListAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	resultJSON, numRecords, err := e.db.getLegalBasisRecords()
	if err != nil {
		ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	//log.Printf("Total count of rows: %d\n", numRecords)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}
