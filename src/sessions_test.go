package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpCreateSession(sid string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/" + sid
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetSession(recordTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/" + recordTOKEN
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserSessions(mode string, identity string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/sessions/" + mode + "/" + identity
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserLoginSessions(login string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/login/" + login
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateSessionRecord(t *testing.T) {
	userJSON := `{"login":"alex"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, found := raw["status"]; !found || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create user")
	}
	userTOKEN := raw["token"].(string)
	data := `{"expiration":"1m","cookie":"abcdefg","login":"alex"}`
	sid, _ := uuid.GenerateUUID()
	raw, _ = helpCreateSession(sid, data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	sessionTOKEN := raw["session"].(string)
	raw, _ = helpGetSession(sessionTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	raw, _ = helpGetUserSessions("token", userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	if raw["total"].(float64) != 1 {
		t.Fatalf("wrong number of sessions")
	}
	data2 := `{"expiration":"1m","cookie":"abcdefg2","login":"alex"}`
	sid2, _ := uuid.GenerateUUID()
	raw, _ = helpCreateSession(sid2, data2)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	raw, _ = helpGetUserSessions("login", "alex")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	if raw["total"].(float64) != 2 {
		t.Fatalf("wrong number of sessions")
	}
}

func TestCreateSessionAndSharedRecord(t *testing.T) {
	userJSON := `{"login":"dima"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, found := raw["status"]; !found || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create user")
	}
	userTOKEN := raw["token"].(string)
	data := `{"expiration":"1m","cookie":"abcdefg","secret":"value"}`
	sid, _ := uuid.GenerateUUID()
	raw, _ = helpCreateSession(sid, data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	sessionTOKEN := raw["session"].(string)
	data = fmt.Sprintf(`{"expiration":"1d","session":"%s","fields":"cookie,missing"}`, sessionTOKEN)
	raw, _ = helpCreateSharedRecord("token", userTOKEN, data)
	recordTOKEN := raw["record"].(string)
	fmt.Printf("User record token: %s\n", recordTOKEN)
	raw, _ = helpGetSharedRecord(recordTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get shared record: %s\n", raw["message"])
	}
}

func TestCreateFakeUserSession(t *testing.T) {
	data := `{"expiration":"1d","cookie":"12345"}`
	sid, _ := uuid.GenerateUUID()
	raw, _ := helpCreateSession(sid, data)
	if _, ok := raw["status"]; ok && raw["status"].(string) != "ok" {
		t.Fatalf("Should fail to create session")
	}
}

func TestGetFakeSession(t *testing.T) {
	rtoken, _ := uuid.GenerateUUID()
	raw, _ := helpGetSession(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
	raw, _ = helpGetSession("faketoken")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
}

func TestGetFakeSessions(t *testing.T) {
	rtoken, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserSessions("token", rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
	raw, _ = helpGetUserSessions("faketoken", rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
}
