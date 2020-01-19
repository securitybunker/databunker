package databunker

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func helpCreateSharedRecord(userTOKEN string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/sharedrecord/token/"+userTOKEN
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
	url := "http://localhost:3000/v1/get/"+recordTOKEN
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
	raw, err = helpCreateSharedRecord(userTOKEN, data)

	if status, ok := raw["status"]; ok {
		if status == "error" {
			t.Fatalf("Failed to create shared record: %s\n", raw["message"])
			return
		} else if status == "ok" {
			recordTOKEN = raw["record"].(string)
		}
	}
	if len(recordTOKEN) == 0 {
		t.Fatalf("Failed to retreave user token: %s\n", raw)
	}
	fmt.Printf("User record token: %s\n", recordTOKEN)
	raw, err = helpGetSharedRecord(recordTOKEN)
	if status, ok := raw["status"]; ok {
		if status == "error" {
			t.Fatalf("Failed to get shared record: %s\n", raw["message"])
			return
		}
	}
	helpDeleteUser("token", userTOKEN)
}
