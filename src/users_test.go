package main

import (
	"fmt"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
	jsonpatch "github.com/evanphx/json-patch"
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

func helpGetUserAuditEvents(userTOKEN string, args string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/audit/list/" + userTOKEN + args
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
	userJSON := `{"login":"user1","name":"tom","height":100,"phone":"775566998822","devices":[{"name":"dev1","val":1},{"name":"dev2","val":2}]}`
	raw, _ := helpCreateUser(userJSON)
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
	patchJSON := `[
		{"op": "replace", "path": "/devices/1", "value": {"name":"updated"}},
		{"op": "add", "path": "/devices/0", "value":{"name":"dev3"}},
		{"op": "remove", "path": "/height"}
	]`
	raw, _ = helpChangeUser("token", userTOKEN, patchJSON)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update user")
	}
	raw, _ = helpGetUser("phone", "775566998822")
	userRecord, _ := json.Marshal(raw["data"].(map[string]interface{}))
	//fmt.Printf("get user %v\n", raw)
	//fmt.Printf("user %s\n", string(userRecord))
	afterUpdate := []byte(`{"devices":[{"name":"dev3"},{"name":"dev1","val":1},{"name":"updated"}],"name":"tom","phone":"775566998822"}`)
	if !jsonpatch.Equal(userRecord, afterUpdate) {
		t.Fatalf("Records are different")
	}
	raw, _ = helpChangeUser("phone", "775566998822", `{"login":"parpar1"}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update user")
	}
	raw, _ = helpChangeUser("phone", "775566998822", `{}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
	}
	raw, _ = helpChangeUser("phone", "775566998822", `a=b`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
	}
	raw, _ = helpGetUser("login", "user1")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Lookup by login should fail now")
	}
	raw, _ = helpGetUserAuditEvents(userTOKEN, "?offset=1&limit=1")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get audit event/s\n")
	}
	records := raw["rows"].([]interface{})
	if raw["total"].(float64) != 6 {
		t.Fatalf("Wrong number of audit event/s\n")
	}
	if len(records) != 1 {
		t.Fatalf("Wrong number of audit rows/s\n")
	}
	raw, _ = helpGetUserAuditEvents(userTOKEN, "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get audit event/s\n")
	}
	records = raw["rows"].([]interface{})
	atoken := ""
	for id := range records {
		records0 := records[id].(map[string]interface{})
		atoken = records0["atoken"].(string)
		if len(atoken) == 0 {
			t.Fatalf("Failed to extract atoken\n")
		}
		fmt.Printf("Audit record: %s\n", atoken)
		raw, _ = helpGetUserAuditEvent(atoken)
		if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
			t.Fatalf("Failed to get specific audit event\n")
		}
	}
	oldRootToken := rootToken
	rootToken, _ = uuid.GenerateUUID()
	raw, _ = helpGetUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Lookup by login should fail now")
	}
	raw, _ = helpGetUserAuditEvents(userTOKEN, "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should faile to get audit event/s\n")
	}
	raw, _ = helpGetUserAuditEvent(atoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get specific audit event\n")
	}
	raw, _ = helpDeleteUser("phone", "775566998822")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to delete user\n")
	}
	rootToken = oldRootToken
	helpDeleteUser("phone", "775566998822")
	helpDeleteUser("token", userTOKEN)
	raw, _ = helpGetUser("token", userTOKEN)
	d := raw["data"].(map[string]interface{})
	if len(d) != 0 {
		t.Fatalf("Failed to delete user")
	}
}

func TestGetFakeAuditEvent(t *testing.T) {
	auditTOKEN := "token123"
	raw, _ := helpGetUserAuditEvent(auditTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user audit events")
	}
}

func TestGetFakeAuditEvent2(t *testing.T) {
	auditTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserAuditEvent(auditTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user audit events")
	}
}

func TestDeleteFakeUser(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpDeleteUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to delete fake user")
	}
}

func TestDeleteFakeUser2(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpDeleteUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to delete fake user")
	}
}

func TestDeleteFakeUserBadMode(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpDeleteUser("fake", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to delete fake user")
	}
}

func TestAuditEventsFakeUser(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpGetUserAuditEvents(userTOKEN, "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user audit events")
	}
}

func TestAuditEventsFakeUser2(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserAuditEvents(userTOKEN, "?offset=1&limit=1")
	//if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
	//	t.Fatalf("Should fail to get user audit events")
	//}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Should return empty list of audit events")
	}
}

func TestGetFakeUserToken(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpGetUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user record")
	}
}

func TestGetFakeUserToken2(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUser("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user record")
	}
}

func TestGetFakeUserBadMode(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUser("fake", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user record")
	}
}

func TestUpdateFakeUser(t *testing.T) {
	userTOKEN := "token123"
	raw, _ := helpChangeUser("token", userTOKEN, `{"login":null}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
	}
}

func TestUpdateFakeUser2(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpChangeUser("token", userTOKEN, `{"login":null}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
	}
}

func TestUpdateFakeUserFakeMode(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpChangeUser("fake", userTOKEN, `{"login":null}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
	}
}

func TestUpdateFakeUserFakeEmail(t *testing.T) {
	raw, _ := helpChangeUser("email", "fake1234@fake1234.com", `{"login":null}`)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to update user")
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

func TestCreateUserEmptyBody(t *testing.T) {
	data := "{}"
	raw, _ := helpCreateUser(data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user")
	}
}

func TestCreateUserDupLogin(t *testing.T) {
	data := `{"login":"dup","name":"dup"}`
	raw, _ := helpCreateUser(data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create dup1 user")
	}
	data = `{"login":"dup","name":"dup2"}`
	raw, _ = helpCreateUser(data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user")
	}
}

func TestCreateUserDupEmail(t *testing.T) {
	data := `{"email":"dup@dupdup.com","name":"dup"}`
	raw, _ := helpCreateUser(data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create dup1 user")
	}
	data = `{"email":"dup@dupdup.com","name":"dup2"}`
	raw, _ = helpCreateUser(data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user")
	}
}

func TestCreateUserDupPhone(t *testing.T) {
	data := `{"phone":"334455667788","name":"dup"}`
	raw, _ := helpCreateUser(data)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create dup1 user")
	}
	data = `{"phone":"334455667788","name":"dup2"}`
	raw, _ = helpCreateUser(data)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user")
	}
}

func TestCreateUserBadPOST(t *testing.T) {
	url := "http://localhost:3000/v1/user"
	data := "name=user6&job=developer&email=user6@user6.com"
	request := httptest.NewRequest("POST", url, strings.NewReader(data))
	request.Header.Set("X-Bunker-Token", rootToken)
	raw, _ := helpServe(request)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to create user")
	}
}

func TestCreateUserEmptyXToken2(t *testing.T) {
	//e.conf.Generic.CreateUserWithoutAccessToken = true
	url := "http://localhost:3000/v1/user"
	data := "name=user8&job=developer&email=user8@user8.com"
	request := httptest.NewRequest("POST", url, strings.NewReader(data))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	raw, _ := helpServe2(request)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Should fail to create user")
	}
}
