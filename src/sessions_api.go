package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) newSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uuidCode := ps.ByName("uuidcode")
	event := audit("create new session", uuidCode)
	defer func() { event.submit(e.db) }()
}
