package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func helpCreateUser(userJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/user"
	request := httptest.NewRequest("POST", url, strings.NewReader(userJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpCreateUser2(userDATA string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/user"
	request := httptest.NewRequest("POST", url, strings.NewReader(userDATA))
	request.Header.Set("X-Bunker-Token", rootToken)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return helpServe2(request)
}

func helpChangeUser(mode string, userTOKEN string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/user/" + mode + "/" + userTOKEN
	request := httptest.NewRequest("PUT", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUser(index string, indexValue string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/user/" + index + "/" + indexValue
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpDeleteUser(index string, indexValue string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/user/" + index + "/" + indexValue
	request := httptest.NewRequest("DELETE", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserAuditEvents(userTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/audit/list/" + userTOKEN
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserAuditEvent(atoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/audit/get/" + atoken
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateUpdateUser(t *testing.T) {

	userJSON := `{"login":"user1","name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}}}`

	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	var userTOKEN string
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN = raw["token"].(string)
	raw, _ = helpGetUser("login", "user1")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	if raw["token"].(string) != userTOKEN {
		t.Fatalf("Wrong user token")
	}
	raw, _ = helpChangeUser("token", userTOKEN, `{"login":null}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update user")
	}
	raw, _ = helpGetUser("login", "user1")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Lookup by login should fail now")
	}
	raw2, _ := helpGetUserAuditEvents(userTOKEN)
	if raw2["status"].(string) != "ok" {
		t.Fatalf("Failed to get audit event/s\n")
	}
	if raw2["total"].(float64) != 3 {
		t.Fatalf("Wrong number of audit event/s\n")
	}
	records := raw2["rows"].([]interface{})
	records0 := records[0].(map[string]interface{})
	atoken := records0["atoken"].(string)
	if len(atoken) == 0 {
		t.Fatalf("Failed to extract atoken\n")
	}
	fmt.Printf("Audit record: %s\n", atoken)
	raw3, _ := helpGetUserAuditEvent(atoken)
	if raw3["status"].(string) != "ok" {
		t.Fatalf("Failed to get specific audit event\n")
	}
	helpDeleteUser("token", userTOKEN)
	raw4, _ := helpGetUser("token", userTOKEN)
	d := raw4["data"].(map[string]interface{})
	if len(d) != 0 {
		t.Fatalf("Failed to delete user")
	}
}

func TestGetFakeUserToken(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpGetUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user record")
	}
}

func TestCreateUser2(t *testing.T) {
	data := "name=user2&job=developer&email=user2@user2.com"
	raw, _ := helpCreateUser2(data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	raw, _ = helpGetUser("email", "user2@user2.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Lookup by email should fail now")
	}
	d := raw["data"].(map[string]interface{})
	if _, ok := d["email"]; !ok || d["email"].(string) != "user2@user2.com" {
		t.Fatalf("Wrong email address")
	}
}
