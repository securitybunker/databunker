package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestCreateAPIUser(t *testing.T) {
	masterKey, _ := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	db, _ := newDB(masterKey, nil)
	var cfg Config
	e := mainEnv{db, cfg}

	rootToken, err := e.db.getRootToken()
	if err != nil {
		t.Fatalf("Failed to retreave root token: %s\n", err)
	}
	userJSON := `{"login":"abcdefg","name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}},"admin":true}`

	request := httptest.NewRequest("POST", "/user", strings.NewReader(userJSON))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)
	//var resp http.ResponseWriter
	rr := httptest.NewRecorder()
	var ps httprouter.Params
	e.userNew(rr, request, ps)

	//fmt.Printf("After create user------------------\n%s\n\n\n", rr.Body)
	var raw map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	if err != nil {
		t.Fatalf("Failed to parse json response on user create: %s\n", err)
	}
	var userTOKEN string
	if status, ok := raw["status"]; ok {
		if status == "error" {
			if strings.HasPrefix(raw["message"].(string), "duplicate") {
				_, userTOKEN, _ = e.db.getUserIndex("abcdefg", "login")
				fmt.Printf("user already exists: %s\n", userTOKEN)
			} else {
				t.Fatalf("Failed to create user: %s\n", raw["message"])
				return
			}
		} else if status == "ok" {
			userTOKEN = raw["token"].(string)
		}
	}
	if len(userTOKEN) == 0 {
		t.Fatalf("Failed to parse user UUID")
	}
	p2 := httprouter.Param{"token", userTOKEN}
	ps2 := []httprouter.Param{p2}

	pars := `{"expiration":"1d","fields":"uuid,name,pass,k1,k2.f3"}`
	request = httptest.NewRequest("POST", "/user", strings.NewReader(pars))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)
	//var resp http.ResponseWriter
	rr = httptest.NewRecorder()
	e.userNewToken(rr, request, ps2)
	//fmt.Printf("after create token------------------\n%s\n\n\n", rr.Body)
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	if err != nil {
		fmt.Printf("Failed to parse json response on user create: %s\n", err)
	}
	tokenUUID := ""
	if status, ok := raw["status"]; ok {
		if status == "error" {
			t.Fatalf("Failed to create user token: %s\n", raw["message"])
			return
		} else if status == "ok" {
			tokenUUID = raw["xtoken"].(string)
		}
	}
	if len(tokenUUID) == 0 {
		t.Fatalf("Failed to retreave user token: %s\n", rr.Body)
	}
	fmt.Printf("User token: %s\n", tokenUUID)

	request = httptest.NewRequest("GET", "/user", nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)
	//var resp http.ResponseWriter
	rr = httptest.NewRecorder()

	p3 := httprouter.Param{"xtoken", tokenUUID}
	ps3 := []httprouter.Param{p3}
	e.userCheckToken(rr, request, ps3)
	fmt.Printf("get by token------------------\n%s\n\n\n", rr.Body)
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	if err != nil {
		fmt.Printf("Failed to parse json response on user create: %s\n", err)
	}

	request = httptest.NewRequest("DELETE", "/user", nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)
	//var resp http.ResponseWriter
	rr = httptest.NewRecorder()
	p4 := httprouter.Param{"code", userTOKEN}
	p5 := httprouter.Param{"index", "token"}
	ps4 := []httprouter.Param{p4, p5}
	e.userDelete(rr, request, ps4)
	fmt.Printf("after userDelete------------------\n%s\n\n\n", rr.Body)
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	if err != nil {
		fmt.Printf("Failed to parse json response on user create: %s\n", err)
	}
}
