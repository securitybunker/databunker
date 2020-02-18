package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func helpUserLogin(mode string, address string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/login/" + mode + "/" + address
	request := httptest.NewRequest("GET", url, nil)
	//request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpUserLoginEnter(mode string, address string, code string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/enter/" + mode + "/" + address + "/" + code
	request := httptest.NewRequest("GET", url, nil)
	//request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserRequests() (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/requests"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpApproveUserRequest(rtoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/request/" + rtoken
	request := httptest.NewRequest("POST", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpGetUserRequest(rtoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/request/" + rtoken
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func helpCancelUserRequest(rtoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/request/" + rtoken
	request := httptest.NewRequest("DELETE", url, nil)
	request.Header.Set("X-Bunker-Token", rootToken)
	return helpServe(request)
}

func TestUserLoginDelete(t *testing.T) {
	email := "test@paranoidguy.com"
	jsonData := `{"email":"test@paranoidguy.com","phone":"22346622","fname":"Yuli","lname":"Str","tz":"323xxxxx","password":"123456","address":"Y-d habanim 7","city":"Petah-Tiqva","btest":true,"numtest":123,"testnul":null}`
	raw, err := helpCreateUser(jsonData)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	status := raw["status"].(string)
	if status == "error" {
		if strings.Contains(raw["message"].(string), "duplicate") {
			raw, _ = helpGetUser("email", email)
		} else {
			t.Fatalf("Failed to create user: %s", raw["message"].(string))
		}
	}
	userTOKEN := raw["token"].(string)
	raw, _ = helpUserLogin("email", email)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user login: %s", raw["message"].(string))
	}
	/*
		userBson, err := e.db.lookupUserRecordByIndex("email", email, e.conf)
		if userBson == nil || err != nil {
			t.Fatalf("Failed to lookupUserRecordByIndex")
		}
		tmpCode := int32(0)
		if _, ok := userBson["tempcode"]; ok {
			tmpCode = userBson["tempcode"].(int32)
		}
	*/
	raw, _ = helpUserLoginEnter("email", email, "4444") //strconv.Itoa(int(tmpCode)))
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user login: %s", raw["message"].(string))
	}
	xtoken := raw["xtoken"].(string)
	fmt.Printf("User login *** xtoken: %s\n", xtoken)
	oldRootToken := rootToken
	rootToken = xtoken
	raw, _ = helpAcceptConsent("token", userTOKEN, "contract", "")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	raw, _ = helpWithdrawConsent("token", userTOKEN, "contract")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to accept on consent")
	}
	helpAcceptConsent("token", userTOKEN, "contract2", "")
	helpWithdrawConsent("token", userTOKEN, "contract2")
	raw, _ = helpChangeUser("token", userTOKEN, `{"login":null}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update user")
	}
	raw, _ = helpCreateUserApp(userTOKEN, "testapp", `{"custom":1}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create app: testapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, "testapp", `{"custom2":"abc"}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update app: testapp")
	}
	raw, _ = helpCreateUserApp(userTOKEN, "testapp2", `{"custom":1}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create app: testapp")
	}
	raw, _ = helpUpdateUserApp(userTOKEN, "testapp2", `{"custom2":"abc"}`)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to update app: testapp")
	}
	raw, _ = helpGetUserAppList(userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to get user app list with user xtoken\n")
	}
	fmt.Printf("apps: %s\n", raw["apps"])
	// user asks to forget-me
	raw, _ = helpDeleteUser("token", userTOKEN)
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to delete user")
	}
	if raw["result"].(string) != "request-created" {
		t.Fatalf("Wrong status. It should be: request-created")
	}
	rtoken0 := raw["rtoken"].(string)
	raw, _ = helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw["apps"])

	rootToken = oldRootToken
	// get user requests
	raw, _ = helpGetUserRequests()
	if raw["total"].(float64) != 5 {
		t.Fatalf("Wrong number of user requests for admin to approve/reject/s\n")
	}
	records := raw["rows"].([]interface{})
	for id := range records {
		records0 := records[id].(map[string]interface{})
		action := records0["action"].(string)
		rtoken := records0["rtoken"].(string)
		if len(rtoken) == 0 {
			t.Fatalf("Failed to extract request token\n")
		}
		if action == "forget-me" {
			if rtoken != rtoken0 {
				t.Fatalf("Rtoken0 is wrong\n")
			}
			fmt.Printf("** User request record: %s\n", rtoken)
		}
		raw8, _ := helpGetUserRequest(rtoken)
		if raw8["status"].(string) != "ok" {
			t.Fatalf("Failed to retrieve user request")
		}
		if action == "consent-withdraw" {
			brief := records0["brief"].(string)
			if brief == "contract" {
				helpApproveUserRequest(rtoken)
			} else {
				helpCancelUserRequest(rtoken)
			}
		} else {
			helpApproveUserRequest(rtoken)
			raw9, _ := helpCancelUserRequest(rtoken)
			if raw9["status"].(string) != "error" {
				t.Fatalf("Cancel request should fail here")
			}
		}
	}

	// user should be deleted now
	raw10, _ := helpGetUserAppList(userTOKEN)
	if len(raw10["apps"].([]interface{})) != 0 {
		t.Fatalf("Apps list shoud be empty\n")
	}
}

func TestBadLogin(t *testing.T) {
	userJSON := `{"login":"user10","email":"user10@user10.com","phone":"8855667788"}`
	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("Error in user creation: %s", err)
	}
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Failed to create user")
	}
	//userTOKEN := raw["token"].(string)
	raw, _ = helpUserLogin("login", "user10")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to login user")
	}
	raw, _ = helpUserLogin("email", "user10@user10.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Fail to login user")
	}
	raw, _ = helpUserLogin("phone", "8855667788")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Fail to login user")
	}
	raw, _ = helpUserLoginEnter("login", "user10", "abc1234")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to login user")
	}
	raw, _ = helpUserLoginEnter("email", "user10@user10.com", "abc1234")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to login user")
	}
}

func TestFakeLogin(t *testing.T) {
	raw, _ := helpUserLogin("email", "user-fake-11@userfake11.com")
	if _, ok := raw["status"]; !ok || raw["status"].(string) != "ok" {
		t.Fatalf("Should be ok for not-existing users")
	}
}

func TestFakeLoginEnter(t *testing.T) {
	raw, _ := helpUserLoginEnter("email", "user-fake-11@userfake11.com", "abc1234")
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to login enter")
	}
}

func TestGetFakeRequest(t *testing.T) {
	rtoken, _ := uuid.GenerateUUID()
	raw, _ := helpGetUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get fake request")
	}
	raw, _ = helpApproveUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to approve fake request")
	}
	raw, _ = helpCancelUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Shoud faile to cancel request")
	}
}

func TestGetFakeRequestToken(t *testing.T) {
	rtoken := "faketoken"
	raw, _ := helpGetUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to get fake request")
	}
	raw, _ = helpApproveUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Should fail to approve fake request")
	}
	raw, _ = helpCancelUserRequest(rtoken)
	if _, ok := raw["status"]; ok && raw["status"].(string) == "ok" {
		t.Fatalf("Shoud faile to cancel request")
	}
}
