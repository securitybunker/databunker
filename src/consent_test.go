package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpAcceptAgreement(mode string, identity string, brief string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/agreement/" + brief + "/" + mode + "/" + identity
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpWithdrawAgreement(mode string, identity string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/agreement/" + brief + "/" + mode + "/" + identity
	request := httptest.NewRequest("DELETE", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserAgreement(mode string, identity string, brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/agreement/" + brief + "/" + mode + "/" + identity
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetAllUserAgreements(mode string, identity string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/agreements/" + mode + "/" + identity
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

/*
func helpGetAllUsersByBrief(brief string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/consents/" + brief
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}
*/

func helpGetLBasis() (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/lbasis"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpCreateLBasis(brief string, dataJSON string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/lbasis/" + brief
	request := httptest.NewRequest("POST", url, strings.NewReader(dataJSON))
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestConsentCreateWithdraw(t *testing.T) {
	raw, _ := helpGetLBasis()
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get all brief codes")
	}
	raw, _ = helpGetAllUserAgreements("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 0 {
		t.Fatalf("Wrong number of user consents")
	}
	raw, _ = helpCreateLBasis("test0", `{"basistype":"consent","usercontrol":true}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create consent")
	}
	helpCreateLBasis("test1", `{"basistype":"consent","usercontrol":true}`)
	brief := "test1"
	raw, _ = helpAcceptAgreement("email", "moshe@moshe-int.com", "test0", `{"expiration":"10m","starttime":0}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpAcceptAgreement("phone", "12345678", "test0", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpGetAllUserAgreements("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 1 {
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
	raw, _ = helpGetAllUserAgreements("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 1 {
		t.Fatalf("Wrong number of user consents")
	}
	raw, _ = helpGetAllUserAgreements("token", userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	raw, _ = helpAcceptAgreement("token", userTOKEN, brief, "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpGetUserAgreement("token", userTOKEN, brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consent")
	}
	raw, _ = helpGetUserAgreement("email", "moshe@moshe-int.com", brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consent")
	}
	record := raw["data"].(map[string]interface{})
	if record["brief"].(string) != brief {
		t.Fatalf("Wrong consent brief value")
	}
	raw, _ = helpWithdrawAgreement("email", "moshe@moshe-int.com", brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to withdraw consent")
	}
	raw, _ = helpWithdrawAgreement("token", userTOKEN, brief)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to withdraw consent")
	}
	raw, _ = helpGetAllUserAgreements("email", "moshe@moshe-int.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user consents")
	}
	if raw["total"].(float64) != 2 {
		t.Fatalf("Wrong number of consents")
	}
	raw, _ = helpGetLBasis()
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get all briefs")
	}
	if raw["total"].(float64) != 4 {
		t.Fatalf("Wrong number of briefs")
	}
	/*
		raw, _ = helpGetAllUsersByBrief(brief)
		if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
			t.Fatalf("Failed to get user consents")
		}
		if raw["total"].(float64) != 1 {
			t.Fatalf("Wrong number of briefs")
		}
	*/
}

/*
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
*/

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
	raw, _ = helpGetUserAgreement("token", userTOKEN, "kolhoz")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}

func TestGetFakeUserConsents(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserAgreement("token", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserAgreement("token", userTOKEN, "ali$baba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserAgreement("fake", userTOKEN, "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserAgreement("token", "faketoken", "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
	raw, _ = helpGetUserAgreement("email", "fakeemail222@fakeemail.com", "alibaba")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get user consent")
	}
}

func TestAcceptConsentEmail(t *testing.T) {
	raw, _ := helpAcceptAgreement("email", "aaa@bb.com", "core-send-email-on-login", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept consent")
	}
}

func TestAcceptConsentFakeUser(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpAcceptAgreement("token", userTOKEN, "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptAgreement("fakemode", "aaa@bb.com", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptAgreement("email", "aaa@bb.com", "br$ief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptAgreement("token", "faketoken", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptAgreement("login", "blahblah", "brief", "")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpAcceptAgreement("phone", "112234889966", "brief", "a=b")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept on consent")
	}
}

func TestWithdrawConsentBadUser(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpWithdrawAgreement("token", userTOKEN, "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawAgreement("token", "badtoken", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawAgreement("fakemode", "aaa@bb.com", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawAgreement("login", "blahblah", "brief")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
	raw, _ = helpWithdrawAgreement("email", "aaa@bb.com", "bri$ef")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to accept consent")
	}
}

func TestGetAllUserConsentsFake(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	raw, _ := helpGetAllUserAgreements("token", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserAgreements("faketoken", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserAgreements("login", userTOKEN)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
	raw, _ = helpGetAllUserAgreements("token", "faketoken")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get all user consents")
	}
}
