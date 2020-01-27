package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func helpCreateUserApp(userTOKEN string, appName string, appJSON string) (map[string]interface{}, error) {
	request := httptest.NewRequest("POST", "http://localhost:3000/v1/userapp/token/"+userTOKEN+"/"+appName, strings.NewReader(appJSON))
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpUpdateUserApp(userTOKEN string, appName string, appJSON string) (map[string]interface{}, error) {
	request := httptest.NewRequest("PUT", "http://localhost:3000/v1/userapp/token/"+userTOKEN+"/"+appName, strings.NewReader(appJSON))
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserApp(userTOKEN string, appName string) (map[string]interface{}, error) {
	request := httptest.NewRequest("GET", "http://localhost:3000/v1/userapp/token/"+userTOKEN+"/"+appName, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpDeleteUserApp(userTOKEN string, appName string) (map[string]interface{}, error) {
	request := httptest.NewRequest("DELETE", "http://localhost:3000/v1/userapp/token/"+userTOKEN+"/"+appName, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserAppList(userTOKEN string) (map[string]interface{}, error) {
	request := httptest.NewRequest("GET", "http://localhost:3000/v1/userapp/token/"+userTOKEN, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetAppList() (map[string]interface{}, error) {
	request := httptest.NewRequest("GET", "http://localhost:3000/v1/userapps", nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestCreateUserApp(t *testing.T) {

	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}}}`

	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping":"done"}`
	appName := "shipping"
	raw2, err := helpCreateUserApp(userTOKEN, appName, appJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if raw2["status"] != "ok" {
		t.Fatalf("Failed to create userapp: %s\n", raw2["message"])
	}
	appJSON = `{"like":"yes"}`
	raw3, err := helpUpdateUserApp(userTOKEN, appName, appJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if raw3["status"] != "ok" {
		t.Fatalf("Failed to update userapp: %s\n", raw3["message"])
		return
	}
	raw4, err := helpGetUserApp(userTOKEN, appName)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if raw4["status"] != "ok" {
		t.Fatalf("Failed to get userapp: %s\n", raw4["message"])
		return
	}
	raw5, err := helpGetUserAppList(userTOKEN)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if raw5["status"] != "ok" {
		t.Fatalf("Failed to get userapp: %s\n", raw5["message"])
		return
	}
	raw6, err := helpGetAppList()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if raw6["status"] != "ok" {
		t.Fatalf("Failed to get userapp: %s\n", raw6["message"])
		return
	}
}
