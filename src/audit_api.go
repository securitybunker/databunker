package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) getAuditEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userTOKEN := ps.ByName("token")
	event := audit("view audit events", userTOKEN, "token", userTOKEN)
	defer func() { event.submit(e.db) }()
	//fmt.Println("error code")
	if enforceUUID(w, userTOKEN, event) == false {
		return
	}
	if e.enforceAuth(w, r, event) == false {
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
	resultJSON, counter, err := e.db.getAuditEvents(userTOKEN, offset, limit)
	if err != nil {
		returnError(w, r, "internal error", 405, err, event)
		return
	}
	fmt.Printf("Total count of events: %d\n", counter)
	//fmt.Fprintf(w, "<html><head><title>title</title></head>")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	str := fmt.Sprintf(`{"total":%d,"rows":%s}`, counter, resultJSON)
	w.Write([]byte(str))
}
