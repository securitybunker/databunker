package utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/securitybunker/databunker/src/audit"
)

func ReturnError(w http.ResponseWriter, r *http.Request, message string, code int, err error, event *audit.AuditEvent) {
	log.Printf("[%d] %s %s -> Return error\n", code, r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"status":"error","message":%q}`, message)
	if event != nil {
		event.Status = "error"
		event.Msg = message
		if err != nil {
			event.Debug = err.Error()
			log.Printf("Generate error response: %s, Error: %s\n", message, err.Error())
		} else {
			log.Printf("Generate error response: %s\n", message)
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
