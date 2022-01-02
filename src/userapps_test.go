package main

import (
	"net/http/httptest"
	"encoding/json"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
	jsonpatch "github.com/evanphx/json-patch"
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

func TestCreateUserAppOnly(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"},"login":"userapp1"}`
	raw, _ := helpCreateUser(userJSON)
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping":"done","height":100,"devices":[{"name":"dev1","val":1},{"name":"dev2","val":2}]}`
	appName := "testapp"
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	appJSON = `{"like":"yes"}`
	raw, _ = helpUpdateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	patchJSON := `[
		{"op": "replace", "path": "/devices/1/name", "value":"updated"},
		{"op": "add", "path": "/devices/0", "value":{"name":"dev3"}},
		{"op": "remove", "path": "/height"}
	]`
	raw, _ = helpUpdateUserApp(userTOKEN, appName, patchJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	raw, _ = helpGetUserApp(userTOKEN, appName)
	userappRecord, _ := json.Marshal(raw["data"].(map[string]interface{}))
	//fmt.Printf("get user %v\n", raw)
	//fmt.Printf("user %s\n", string(userRecord))
	afterUpdate := []byte(`{"devices":[{"name":"dev3"},{"name":"dev1","val":1},{"name":"updated","val":2}],"like":"yes","shipping":"done"}`)
	if !jsonpatch.Equal(userappRecord, afterUpdate) {
		t.Fatalf("Records are different")
	}
	raw, _ = helpGetUserAppList(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get userapp")
	}
	raw, _ = helpGetAppList()
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get userapp list")
	}
}

func TestCreateUserUpdateAppBadData(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"},"login":"userapp2"}`
	raw, _ := helpCreateUser(userJSON)
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping2":"done"}`
	appName := "shipping"
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, "a:b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, "{}")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, "app!123", `{"a":"b"}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, "fakeapp", `{"a":"b"}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpUpdateUserApp("faketoken", appName, `{"a":"b"}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, `{"a":"b"}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	raw, _ = helpGetUserApp(userTOKEN, "fakeapp")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get app detailes for user")
	}
	raw, _ = helpGetUserApp(userTOKEN, "app!name")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get app detailes for user")
	}
}

func TestCreateUserAppResetData(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"},"login":"userapp3"}`
	raw, _ := helpCreateUser(userJSON)
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping":"done"}`
	appName := "shipping"
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, `{"shipping":true}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, appName, `{"shipping":null}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update userapp")
	}
	raw, _ = helpGetUserApp(userTOKEN, appName)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get app detailes for user")
	}
	data := raw["data"].(map[string]interface{})
	if len(data) != 0 {
		t.Fatalf("Expected empty data")
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
	appName := "ship!ping"
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

func TestGetAppListFakeUser(t *testing.T) {
	raw, _ := helpGetUserAppList("faketoken")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user app list")
	}
}

func TestGetFakeApp(t *testing.T) {
	raw, _ := helpGetUserApp("fakeuser", "fakeapp")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get app detailes for user")
	}
}

func TestUserAppAnonymouse(t *testing.T) {
	userJSON := `{"name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1"},"login":"userapp4"}`
	raw, _ := helpCreateUser(userJSON)
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping2":"done"}`
	appName := "shipping"
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	oldRootToken := rootToken
	rootToken, _ = uuid.GenerateUUID()
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create userapp")
	}
	appJSON = `{"like":"yes"}`
	raw, _ = helpUpdateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update userapp")
	}
	raw, _ = helpGetUserApp(userTOKEN, appName)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get userapp")
	}
	raw, _ = helpGetUserAppList(userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get userapp")
	}
	raw, _ = helpGetAppList()
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get userapp list")
	}
	rootToken = oldRootToken
}

func TestCreateUserAppShared(t *testing.T) {
	userJSON := `{"login":"tdkuser"}`
	raw, _ := helpCreateUser(userJSON)
	userTOKEN := raw["token"].(string)
	appJSON := `{"shipping2":"done"}`
	appName := "shipping"
	raw, _ = helpCreateUserApp(userTOKEN, appName, appJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create userapp")
	}
	data := `{"expiration":"1d","app":"shipping","fields":"shipping2"}`
	raw, _ = helpCreateSharedRecord("token", userTOKEN, data)
	recordTOKEN := raw["record"].(string)
	//fmt.Printf("User record token: %s\n", recordTOKEN)
	raw, _ = helpGetSharedRecord(recordTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get shared record: %s\n", raw["message"])
	}
}
