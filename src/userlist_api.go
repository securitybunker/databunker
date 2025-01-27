package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/securitybunker/databunker/src/utils"
)

func (e mainEnv) userList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.EnforceAdmin(w, r, nil) == "" {
		return
	}
	if e.conf.Generic.ListUsers == false {
		utils.ReturnError(w, r, "access denied", 403, nil, nil)
		return
	}
	var offset int32 = 0
	var limit int32 = 10
	args := r.URL.Query()
	if value, ok := args["offset"]; ok {
		offset = utils.Atoi(value[0])
	}
	if value, ok := args["limit"]; ok {
		limit = utils.Atoi(value[0])
	}
	resultJSON, counter, _ := e.db.getUsers(offset, limit)
	log.Printf("Total count of events: %d\n", counter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	if counter == 0 {
		str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":[]}`, counter)
		w.Write([]byte(str))
	} else {
		str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, counter, resultJSON)
		w.Write([]byte(str))
	}
}
