package utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/securitybunker/databunker/src/audit"
)

func ReturnError(w http.ResponseWriter, r *http.Request, message string, code int, err error, event *audit.AuditEvent) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"status":"error","message":%q}`, message)
	if event != nil {
		event.Status = "error"
		event.Msg = message
		if err != nil {
			event.Debug = err.Error()
			log.Printf("ERROR [%d] %s %s -> %s : %s", code, r.Method, r.URL.Path, message, event.Debug)
		} else {
			log.Printf("ERROR [%d] %s %s -> %s", code, r.Method, r.URL.Path, message)
		}
	}
	//http.Error(w, http.StatusText(405), 405)
}

func EnforceUUID(w http.ResponseWriter, uuidCode string, event *audit.AuditEvent) bool {
	if CheckValidUUID(uuidCode) == false {
		//fmt.Printf("405 bad uuid in : %s\n", uuidCode)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintf(w, `{"status":"error","message":"bad uuid"}`)
		if event != nil {
			event.Status = "error"
			event.Msg = "bad uuid"
		}
		return false
	}
	return true
}
