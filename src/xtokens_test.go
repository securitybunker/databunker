package databunker

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	//"strconv"
	"strings"
	"testing"
)

func helpCreateUserLogin(mode string, address string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/login/"+mode+"/"+address
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	//request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpCreateUserLoginEnter(mode string, address string, code string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/enter/"+mode+"/"+address+"/"+code
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	//request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestUserLogin(t *testing.T) {
	email := "test@paranoidguy.com"
	jsonData := `{"email":"test@paranoidguy.com","phone":"22346622","fname":"Yuli","lname":"Str","tz":"323xxxxx","password":"123456","address":"Y-d habanim 7","city":"Petah-Tiqva","btest":true,"numtest":123,"testnul":null}`
	raw, err := helpCreateUser(jsonData)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	status := raw["status"].(string)
	if status == "error" {
		if strings.Contains(raw["message"].(string), "duplicate") {
			raw, err = helpGetUser("email", email)
		} else {
			t.Fatalf("Failed to create user: %s", raw["message"].(string))
		}
	}
	userTOKEN := raw["token"].(string)
	raw2, err := helpCreateUserLogin("email", email) 
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
	raw3, err := helpCreateUserLoginEnter("email", email, "4444") //strconv.Itoa(int(tmpCode))) 
	status = raw3["status"].(string)
	if status == "error" {
		t.Fatalf("Failed to create user login: %s", raw3["message"].(string))
	}
	xtoken := raw3["xtoken"].(string)
	fmt.Printf("User login token: %s\n", xtoken)
	raw4, err := helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw4["apps"])
	helpCreateUserApp(userTOKEN, "qq", `{"custom":1}`)
	raw5, err := helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw5["apps"])
}