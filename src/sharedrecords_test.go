package databunker

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpCreateSharedRecord(userTOKEN string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/sharedrecord/token/" + userTOKEN
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetSharedRecord(recordTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/get/" + recordTOKEN
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestCreateSharedRecord(t *testing.T) {
	userJSON := `{"login":"abcdefg","name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}},"admin":true}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	var userTOKEN string
	var recordTOKEN string
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
	raw, _ = helpCreateSharedRecord(userTOKEN, data)

	if status, ok := raw["status"]; ok {
		if status == "error" {
			t.Fatalf("Failed to create shared record: %s\n", raw["message"])
			return
		} else if status == "ok" {
			recordTOKEN = raw["record"].(string)
		}
	}
	if len(recordTOKEN) == 0 {
		t.Fatalf("Failed to retrieve user token: %s\n", raw)
	}
	fmt.Printf("User record token: %s\n", recordTOKEN)
	raw, _ = helpGetSharedRecord(recordTOKEN)
	if status, ok := raw["status"]; ok {
		if status == "error" {
			t.Fatalf("Failed to get shared record: %s\n", raw["message"])
			return
		}
	}
	helpDeleteUser("token", userTOKEN)
}

func TestFailCreateSharedRecord(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	data := `{"expiration":"1d","fields":"uuid,name,pass,k1"}`
	raw, _ := helpCreateSharedRecord(userTOKEN, data)

	if raw["status"].(string) == "ok" {
		t.Fatalf("Created shared record for non-existing user\n")
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
