package main

import (
	"net/http/httptest"
	"testing"
)

func helpGetExpStatus(utoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/exp/status/token/" + utoken
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpStartExp(utoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/exp/start/token/" + utoken
	request := httptest.NewRequest("POST", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestExpStart(t *testing.T) {
	userJSON := `{"login":"william"}`
	raw, _ := helpCreateUser(userJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
	if _, ok := raw["expstatus"]; !ok || raw["expstatus"].(string) != "" {
		t.Fatalf("Failed to get exp status")
	}
	raw, _ = helpStartExp(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to start expiration")
	}
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
	if _, ok := raw["expstatus"]; !ok || raw["expstatus"].(string) != "wait" {
		t.Fatalf("Exp status is wrong")
	}
}

