package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (e mainEnv) newSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := ps.ByName("token")
	event := audit("create new session", token, "token", token)
	defer func() { event.submit(e.db) }()
}
