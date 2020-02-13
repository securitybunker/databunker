package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	//uuid "github.com/hashicorp/go-uuid"
)

func helpAcceptConsent(mode string, address string, brief string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
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

func helpWithdrawConsent(mode string, address string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
	request := httptest.NewRequest("DELETE", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserConsent(mode string, address string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestCreateWithdrawConsent(t *testing.T) {
	userJSON := `{"login":"moshe", "email":"moshe@moshe-int.com"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, found := raw["status"]; !found || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create user")
	}
	userTOKEN := raw["token"].(string)
	bief := "test1"
	raw, _ = helpAcceptConsent("email", "moshe@moshe-int.com", bief, "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	raw, _ = helpGetUserConsent("token", userTOKEN, bief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("failed to create session")
	}
	record := raw["data"].(map[string]interface{})
	if record["brief"].(string) != bief {
		t.Fatalf("wrong concent brief value")
	}
}
