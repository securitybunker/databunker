package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpCreateSharedRecord(mode string, identity string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/sharedrecord/" + mode + "/" + identity
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetSharedRecord(recordTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/get/" + recordTOKEN
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateSharedRecord(t *testing.T) {
	userJSON := `{"login":"abcdefg","name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}},"admin":true}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	var userTOKEN string
	if status, ok := raw["status"]; ok {
		if status == "error" {
			if strings.HasPrefix(raw["message"].(string), "duplicate") {
				raw2, _ := helpGetUser("login", "abcdefg")
				userTOKEN = raw2["token"].(string)
			} else {
				t.Fatalf("Failed to create user: %s\n", raw["message"])
				return
			}
		} else if status == "ok" {
			userTOKEN = raw["token"].(string)
		}
	}
	if len(userTOKEN) == 0 {
		t.Fatalf("Failed to parse user token")
	}
	data := `{"expiration":"1d","fields":"uuid,name,pass,k1,k2.f3"}`
	raw, _ = helpCreateSharedRecord("token", userTOKEN, data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create shared record: %s\n", raw["message"])
	}
	recordTOKEN := raw["record"].(string)
	fmt.Printf("User record token: %s\n", recordTOKEN)
	raw, _ = helpGetSharedRecord(recordTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get shared record: %s\n", raw["message"])
	}
	helpDeleteUser("token", userTOKEN)
}

func TestCreateSharedRecordFakeUser(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	data := `{"expiration":"1d","fields":"uuid,name,pass,k1"}`
	raw, _ := helpCreateSharedRecord("token", userTOKEN, data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create shared record for fake user")
	}
}

func TestCreateSharedRecordBadInput(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpCreateSharedRecord("token", userTOKEN, "a=b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create shared record for fake user")
	}
	data := `{"expiration":"1d","fields":"uuid,name,pass,k1"}`
	raw, _ = helpCreateSharedRecord("token", userTOKEN, "a=b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create shared record for fake user")
	}
	raw, _ = helpCreateSharedRecord("faketoken", userTOKEN, data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create shared record for fake user")
	}
	raw, _ = helpCreateSharedRecord("token", "faketoken", data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create shared record for fake user")
	}
}

func TestGetFakeSharedRecord(t *testing.T) {
	rtoken, _ := uuid.GenerateUUID()
	_, err := helpGetSharedRecord(rtoken)
	if err == nil {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
}

/*
func Test_UserAppToken(t *testing.T) {
	masterKey, err := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	db, _ := newDB(masterKey, nil)

	var parsedData userJSON
	parsedData.jsonData = []byte(`{"login":"start","field":"bbb"}`)
	parsedData.loginIdx = "start"
	userTOKEN, err := db.createUserRecord(parsedData, nil)
	fields := "abc"
	expiration := "7d"
	appName := "test"
	userXToken, err := db.generateUserTempXToken(userTOKEN, fields, expiration, appName)
	if err != nil {
		t.Fatalf("Failed to generate user token: %s ", err)
	}
	if userXToken == "" {
		t.Fatalf("Failed to generate user token")
	}
	appName = "test2"
	userXToken, err = db.generateUserTempXToken(userTOKEN, fields, expiration, appName)
	if err == nil {
		t.Fatalf("Using unknown app, should fail.")
	}
	if userXToken != "" {
		t.Fatalf("Should fail to generate user token")
	}
	_, err = db.deleteUserRecord(userTOKEN)
	if err != nil {
		t.Fatalf("Failed to delete user: %s", err)
	}
}
*/
