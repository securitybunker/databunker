package databunker

import (
	"fmt"
	"net/http/httptest"
	//"strconv"
	"strings"
	"testing"
)

func helpCreateUserLogin(mode string, address string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/login/" + mode + "/" + address
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

func helpCreateUserLoginEnter(mode string, address string, code string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/enter/" + mode + "/" + address + "/" + code
	request := httptest.NewRequest("GET", url, nil)
	//request.Header.Set("X-Bunker-Token", rootToken)
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
	raw2, _ := helpCreateUserLogin("email", email)
	status = raw2["status"].(string)
	if status == "error" {
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
	raw3, _ := helpCreateUserLoginEnter("email", email, "4444") //strconv.Itoa(int(tmpCode)))
	if raw3["status"].(string) == "error" {
		t.Fatalf("Failed to create user login: %s", raw3["message"].(string))
	}
	xtoken := raw3["xtoken"].(string)
	fmt.Printf("User login *** xtoken: %s\n", xtoken)
	oldRootToken := rootToken
	rootToken = xtoken
	raw4, _ := helpGetUserAppList(userTOKEN)
	if raw4["status"].(string) == "error" {
		t.Fatalf("Failed to get user app list with user xtoken\n")
	}
	fmt.Printf("apps: %s\n", raw4["apps"])
	// user asks to forget-me
	raw5, _ := helpDeleteUser("token", userTOKEN)
	if raw5["status"].(string) != "ok" {
		t.Fatalf("Failed to delete user")
	}
	if raw5["result"].(string) != "request-created" {
	}
	rtoken0 := raw5["rtoken"].(string)
	rootToken = oldRootToken
	// get user requests
	raw6, _ := helpGetUserRequests()
	if raw6["total"].(float64) != 1 {
		t.Fatalf("Wrong number of audit event/s\n")
	}
	records := raw6["rows"].([]interface{})
	records0 := records[0].(map[string]interface{})
	rtoken := records0["rtoken"].(string)
	if len(rtoken) == 0 {
		t.Fatalf("Failed to extract request token\n")
	}
	if rtoken != rtoken0 {
		t.Fatalf("Rtoken0 is wrong\n")
	}
	fmt.Printf("** User request record: %s\n", rtoken)
	helpCreateUserApp(userTOKEN, "qq", `{"custom":1}`)
	raw7, _ := helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw7["apps"])
	helpApproveUserRequest(rtoken)
	// user should be deleted now
	raw8, _ := helpGetUserAppList(userTOKEN)
	if raw8["apps"] != nil {
		t.Fatalf("Apps shoud be nil\n")
	}
}
