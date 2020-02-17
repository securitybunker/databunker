package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpCreateUserApp(userTOKEN string, appName string, appJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapp/token/" + userTOKEN + "/" + appName
	request := httptest.NewRequest("POST", url, strings.NewReader(appJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpUpdateUserApp(userTOKEN string, appName string, appJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapp/token/" + userTOKEN + "/" + appName
	request := httptest.NewRequest("PUT", url, strings.NewReader(appJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserApp(userTOKEN string, appName string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapp/token/" + userTOKEN + "/" + appName
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpDeleteUserApp(userTOKEN string, appName string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapp/token/" + userTOKEN + "/" + appName
	request := httptest.NewRequest("DELETE", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserAppList(userTOKEN string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapp/token/" + userTOKEN
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetAppList() (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/userapps"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateUserApp(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"}}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping":"done"}`
	appName := "shipping"
	raw, err = helpCreateUserApp(userTOKEN, appName, appJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	appJSON = `{"like":"yes"}`
	raw, err = helpUpdateUserApp(userTOKEN, appName, appJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	raw, err = helpGetUserApp(userTOKEN, appName)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get userapp")
		return
	}
	raw, err = helpGetUserAppList(userTOKEN)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get userapp")
	}
	raw, err = helpGetAppList()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get userapp list")
	}
}

func TestCreateUserUpdateAppBadData(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"}}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping2":"done"}`
	appName := "shipping"
	raw, err = helpCreateUserApp(userTOKEN, appName, appJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, "a:b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should failed to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, "{}")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should failed to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, "app$123", `{"a":"b"}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should failed to update userapp")
	}
}

func TestCreateUserAppFakeToken(t *testing.T) {
	userTOKEN := "token123"
	appJSON := `{"shipping":"done"}`
	appName := "shipping"
	raw, _ := helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user app")
	}
}

func TestCreateUserAppBadAppName(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	appJSON := `{"shipping":"done"}`
	appName := "ship$ping"
	raw, _ := helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user app")
	}
}

func TestCreateUserAppBadData(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	appJSON := `a=b`
	appName := "shipping"
	raw, _ := helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user app")
	}
}

func TestCreateUserAppEmptyData(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	appJSON := ``
	appName := "shipping"
	raw, _ := helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user app")
	}
}
