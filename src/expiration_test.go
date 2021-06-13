package main

import (
	"fmt"
	"time"
	"strings"
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

func helpStartExp2(utoken string, t int32) (map[string]interface{}, error) {
	data := fmt.Sprintf(`{"expiration": %d}`, t)
	url := "http://localhost:3000/v1/exp/start/token/" + utoken
	request := httptest.NewRequest("POST", url, strings.NewReader(data))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpRetainData(exptoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/exp/retain/" + exptoken
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpDeleteData(exptoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/exp/delete/" + exptoken
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpCancelExpiration(utoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/exp/cancel/token/" + utoken
	request := httptest.NewRequest("DELETE", url, nil)
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
	if _, ok := raw["exptime"]; !ok || raw["exptime"].(float64) <= 0 {
		t.Fatalf("Exp endtime is broken")
	}
	exptoken := raw["exptoken"].(string)
	helpRetainData(exptoken)
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
}

func TestExpDel(t *testing.T) {
	userJSON := `{"login":"william2"}`
	raw, _ := helpCreateUser(userJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpStartExp(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to start expiration")
	}
	exptoken := raw["exptoken"].(string)
	helpDeleteData(exptoken)
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
}

func TestExpCancel(t *testing.T) {
	userJSON := `{"login":"william3"}`
	raw, _ := helpCreateUser(userJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpStartExp(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to start expiration")
	}
	//exptoken := raw["exptoken"].(string)
	raw, _ = helpCancelExpiration(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to cancel expiration")
	}
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
}

func TestExpAuto(t *testing.T) {
	userJSON := `{"login":"william4"}`
	now := int32(time.Now().Unix())+1
	fmt.Printf("time %d\n", now)
	raw, _ := helpCreateUser(userJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpStartExp2(userTOKEN, now)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to start expiration")
	}
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
	time.Sleep(2 * time.Second)
	e.expUsers()
	raw, _ = helpGetExpStatus(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get status")
	}
}
