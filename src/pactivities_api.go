package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	//"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) pactivityCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	activity = normalizeBrief(activity)
	if isValidBrief(activity) == false {
		returnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	records, err := getJSONPostMap(r)
	if err != nil {
		returnError(w, r, "failed to decode request body", 405, err, nil)
		return
	}
	defer func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}()

	title := ""
	script := ""
	fulldesc := ""
	legalbasis := ""
	newactivity := ""
	applicableto := ""
	if value, ok := records["title"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			title = value.(string)
		}
	}
	if len(title) == 0 {
		title = activity
	}
	if value, ok := records["script"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			script = value.(string)
		}
	}
	if value, ok := records["fulldesc"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			fulldesc = value.(string)
		}
	}
	if value, ok := records["activity"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			newactivity = value.(string)
		}
	}
	if value, ok := records["applicableto"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			applicableto = value.(string)
		}
	}
	e.db.createProcessingActivity(activity, newactivity, title, script, fulldesc, legalbasis, applicableto)
}

func (e mainEnv) pactivityDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	activity = normalizeBrief(activity)
	if isValidBrief(activity) == false {
		returnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	e.db.deleteProcessingActivity(activity)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) pactivityLink(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	brief := ps.ByName("brief")
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	activity = normalizeBrief(activity)
	if isValidBrief(activity) == false {
		returnError(w, r, "bad activity format", 405, nil, nil)
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
	_, err = e.db.linkProcessingActivity(activity, brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) pactivityUnlink(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	brief := ps.ByName("brief")
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	activity = normalizeBrief(activity)
	if isValidBrief(activity) == false {
		returnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	brief = normalizeBrief(brief)
	if isValidBrief(brief) == false {
		returnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	_, err := e.db.unlinkProcessingActivity(activity, brief)
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) pactivityList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authResult := e.enforceAdmin(w, r)
	if authResult == "" {
		return
	}
	resultJSON, numRecords, err := e.db.listProcessingActivities()
	if err != nil {
		returnError(w, r, "internal error", 405, err, nil)
		return
	}
	fmt.Printf("Total count of rows: %d\n", numRecords)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}
