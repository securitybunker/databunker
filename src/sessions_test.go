package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpCreateSession(userTOKEN string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/token/" + userTOKEN
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

func helpGetSession(recordTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/session/" + recordTOKEN
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserTokenSessions(userTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/token/" + userTOKEN
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserLoginSessions(login string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/session/login/" + login
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
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
	data := `{"expiration":"1m","cookie":"abcdefg"}`
	raw, _ = helpCreateSession(userTOKEN, data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	sessionTOKEN := raw["session"].(string)
	raw, _ = helpGetSession(sessionTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	raw, _ = helpGetUserTokenSessions(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	if raw["total"].(float64) != 1 {
		t.Fatalf("wrong number of sessions")
	}
	data2 := `{"expiration":"1m","cookie":"abcdefg2"}`
	raw, _ = helpCreateSession(userTOKEN, data2)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	raw, _ = helpGetUserLoginSessions("alex")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to get session")
	}
	if raw["total"].(float64) != 2 {
		t.Fatalf("wrong number of sessions")
	}
}

func TestFailCreateSession(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	data := `{"expiration":"1d","cookie":"12345"}`
	raw, _ := helpCreateSession(userTOKEN, data)

	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Created session record for non-existing user\n")
	}
}

func TestGetFakeSession(t *testing.T) {
	rtoken, _ := uuid.GenerateUUID()
	raw, _ := helpGetSession(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to retrieve non-existing record\n")
	}
}
