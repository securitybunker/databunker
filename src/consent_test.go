package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpAcceptConsent(mode string, address string, brief string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpWithdrawConsent(mode string, address string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
	request := httptest.NewRequest("DELETE", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserConsent(mode string, address string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address + "/" + brief
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetAllUserConsents(mode string, address string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consent/" + mode + "/" + address
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetAllUsersByBrief(brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consents/" + brief
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetLBasis() (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/lbasis"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateWithdrawConsent(t *testing.T) {
	raw, _ := helpGetLBasis()
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get all brief codes")
	}
	raw, _ = helpGetAllUserConsents("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Wrong number of user consents")
	}
	brief := "test1"
	raw, _ = helpAcceptConsent("email", "moshe@moshe-int.com", "test0", `{"expiration":"10m","starttime":0}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpAcceptConsent("phone", "12345678", "test0", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpGetAllUserConsents("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 2 {
		t.Fatalf("Wrong number of user consents")
	}
	userJSON := `{"login":"moshe","email":"moshe@moshe-int.com","phone":"12345678"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Wrror in user creation: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpGetAllUserConsents("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 1 {
		t.Fatalf("Wrong number of user consents")
	}
	raw, _ = helpGetAllUserConsents("token", userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	raw, _ = helpAcceptConsent("token", userTOKEN, brief, "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpAcceptConsent("email", "moshe@moshe-int.com", "contract-accept", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent: contract-accept")
	}
	raw, _ = helpGetUserConsent("token", userTOKEN, brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consent")
	}
	raw, _ = helpGetUserConsent("email", "moshe@moshe-int.com", brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consent")
	}
	record := raw["data"].(map[string]interface{})
	if record["brief"].(string) != brief {
		t.Fatalf("Wrong consent brief value")
	}
	raw, _ = helpWithdrawConsent("email", "moshe@moshe-int.com", brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to withdraw consent")
	}
	raw, _ = helpWithdrawConsent("token", userTOKEN, brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to withdraw consent")
	}
	raw, _ = helpGetAllUserConsents("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 3 {
		t.Fatalf("Wrong number of consents")
	}
	raw, _ = helpGetLBasis()
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get all briefs")
	}
	if raw["total"].(float64) != 3 {
		t.Fatalf("Wrong number of briefs")
	}
	raw, _ = helpGetAllUsersByBrief(brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 1 {
		t.Fatalf("Wrong number of briefs")
	}
}

func TestGetFakeBrief(t *testing.T) {
	raw, _ := helpGetAllUsersByBrief("unknown")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Should fail to get all users with this brief")
	}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Wrong number of briefs")
	}
	raw, _ = helpGetAllUsersByBrief("unk$nown")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Should fail to get all users with this brief")
	}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Wrong number of briefs")
	}
}

func TestGetUserUnkConsent(t *testing.T) {
	userJSON := `{"email":"moshe23@mosh23e-int.com","phone":"123586678"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Wrror in user creation: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpGetUserConsent("token", userTOKEN, "kolhoz")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}

func TestGetFakeUserConsents(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserConsent("token", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserConsent("token", userTOKEN, "ali$baba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserConsent("fake", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserConsent("token", "faketoken", "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserConsent("email", "fakeemail222@fakeemail.com", "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}

func TestAcceptConsentEmail(t *testing.T) {
	raw, _ := helpAcceptConsent("email", "aaa@bb.com", "brief", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept consent")
	}
}

func TestAcceptConsentFakeUser(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpAcceptConsent("token", userTOKEN, "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptConsent("fakemode", "aaa@bb.com", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptConsent("email", "aaa@bb.com", "br$ief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptConsent("token", "faketoken", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptConsent("login", "blahblah", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptConsent("phone", "112234889966", "brief", "a=b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept on consent")
	}
}

func TestWithdrawConsentBadUser(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpWithdrawConsent("token", userTOKEN, "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawConsent("token", "badtoken", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawConsent("fakemode", "aaa@bb.com", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawConsent("login", "blahblah", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawConsent("email", "aaa@bb.com", "bri$ef")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
}

func TestGetAllUserConsentsFake(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetAllUserConsents("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserConsents("faketoken", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserConsents("login", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserConsents("token", "faketoken")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
}
