package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/utils"
	//"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) pactivityCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	activity = utils.NormalizeBrief(activity)
	if utils.CheckValidBrief(activity) == false {
		utils.ReturnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	records, err := utils.GetJSONPostMap(r)
	if err != nil {
		utils.ReturnError(w, r, "failed to decode request body", 405, err, nil)
		return
	}
	defer func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}()

	legalbasis := ""
	title := utils.GetStringValue(records["title"])
	if len(title) == 0 {
		title = activity
	}
	script := utils.GetStringValue(records["script"])
	fulldesc := utils.GetStringValue(records["fulldesc"])
	newactivity := utils.GetStringValue(records["activity"])
	applicableto := utils.GetStringValue(records["applicableto"])

	e.db.createProcessingActivity(activity, newactivity, title, script, fulldesc, legalbasis, applicableto)
}

func (e mainEnv) pactivityDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	activity = utils.NormalizeBrief(activity)
	if utils.CheckValidBrief(activity) == false {
		utils.ReturnError(w, r, "bad activity format", 405, nil, nil)
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
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	activity = utils.NormalizeBrief(activity)
	if utils.CheckValidBrief(activity) == false {
		utils.ReturnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	brief = utils.NormalizeBrief(brief)
	if utils.CheckValidBrief(brief) == false {
		utils.ReturnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	exists, err := e.db.checkLegalBasis(brief)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, nil)
		return
	}
	if exists == false {
		utils.ReturnError(w, r, "not found", 405, nil, nil)
		return
	}
	_, err = e.db.linkProcessingActivity(activity, brief)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) pactivityUnlink(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activity := ps.ByName("activity")
	brief := ps.ByName("brief")
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	activity = utils.NormalizeBrief(activity)
	if utils.CheckValidBrief(activity) == false {
		utils.ReturnError(w, r, "bad activity format", 405, nil, nil)
		return
	}
	brief = utils.NormalizeBrief(brief)
	if utils.CheckValidBrief(brief) == false {
		utils.ReturnError(w, r, "bad brief format", 405, nil, nil)
		return
	}
	_, err := e.db.unlinkProcessingActivity(activity, brief)
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) pactivityListAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	resultJSON, numRecords, err := e.db.listProcessingActivities()
	if err != nil {
		utils.ReturnError(w, r, "internal error", 405, err, nil)
		return
	}
	log.Printf("Total count of rows: %d\n", numRecords)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
	w.Write([]byte(str))
}
