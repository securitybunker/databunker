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

func helpGetAllBriefs() (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consents"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestCreateWithdrawConsent(t *testing.T) {
	raw, _ := helpGetAllBriefs()
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
	raw, _ = helpAcceptConsent("email", "moshe@moshe-int.com", "test0", "")
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
	raw, _ = helpGetAllBriefs()
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
		t.Fatalf("Should fal to get all users with this brif")
	}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Wrong number of briefs")
	}
}

func TestGetFakeUserConsents(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserConsent("token", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}

func TestGetFakeUserConsents2(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserConsent("fake", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}
