package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func Test_UserTempToken(t *testing.T) {
	masterKey, err := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	db, _ := newDB(masterKey, nil)

	var parsedData userJSON
	parsedData.jsonData = []byte(`{"login":"start","key1":"bbb","key2":[10,20]}`)
	parsedData.loginIdx = "start"
	userTOKEN, err := db.createUserRecord(parsedData, nil)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	fmt.Printf("user token generated: %s\n", userTOKEN)
	fields := "key1,key2.1"
	expiration := "7d"
	userToken, err := db.generateUserTempXToken(userTOKEN, fields, expiration, "")
	if err != nil {
		t.Fatalf("Failed to generate user token: %s ", err)
	}
	if userToken == "" {
		t.Fatalf("Failed to generate user token")
	}
	_, err = db.deleteUserRecord(userTOKEN)
	if err != nil {
		t.Fatalf("Failed to delete user: %s", err)
	}
}

func Test_UserTempToken2(t *testing.T) {
	masterKey, err := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	db, _ := newDB(masterKey, nil)

	userTOKEN, err := uuid.GenerateUUID()
	fields := "abc"
	expiration := "7d"
	_, err = db.generateUserTempXToken(userTOKEN, fields, expiration, "")
	if err == nil {
		t.Fatalf("Should failed to generate user token")
	}
}

func helpCreateUserXToken(uuidCode string, tokenJSON string) (map[string]interface{}, error) {
	request := httptest.NewRequest("POST",
		"http://localhost:3000/v1/xtoken/"+uuidCode,
		strings.NewReader(tokenJSON))
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestAPIToken(t *testing.T) {
	jsonData := `{"email":"stremovsky@gmail.com","phone":"0524486622","fname":"Yuli","lname":"Stremovsky","tz":"323xxxxx","password":"123456","address":"Y-d habanim 7","city":"Petah-Tiqva","btest":true,"numtest":123,"testnul":null}`
	raw, err := helpCreateUser(jsonData)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			raw, err = helpGetUser("email", "stremovsky@gmail.com")
		} else {
			t.Fatalf("error: %s", err)
		}
	}
	status := raw["status"].(string)
	if status == "error" {
		if strings.Contains(raw["message"].(string), "duplicate") {
			raw, err = helpGetUser("email", "stremovsky@gmail.com")
		} else {
			t.Fatalf("Failed to create user: %s", raw["message"].(string))
		}
	}
	userTOKEN := raw["token"].(string)
	fields := "phone,field1,field2"
	tokenJSON := fmt.Sprintf(`{"fields":"%s","expiration":"1d"}`, fields)
	raw2, err := helpCreateUserXToken(userTOKEN, tokenJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	token := raw2["xtoken"].(string)
	fmt.Printf("**** Result token : %s\n", token)
	raw3, err := helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw3["apps"])
	helpCreateUserApp(userTOKEN, "qq", `{"custom":1}`)
	raw3, err = helpGetUserAppList(userTOKEN)
	fmt.Printf("apps: %s\n", raw3["apps"])
}

func Test_UserAppToken(t *testing.T) {
	masterKey, err := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	db, _ := newDB(masterKey, nil)

	var parsedData userJSON
	parsedData.jsonData = []byte(`{"login":"start","field":"bbb"}`)
	parsedData.loginIdx = "start"
	userTOKEN, err := db.createUserRecord(parsedData, nil)
	fields := "abc"
	expiration := "7d"
	appName := "test"
	userXToken, err := db.generateUserTempXToken(userTOKEN, fields, expiration, appName)
	if err != nil {
		t.Fatalf("Failed to generate user token: %s ", err)
	}
	if userXToken == "" {
		t.Fatalf("Failed to generate user token")
	}
	appName = "test2"
	userXToken, err = db.generateUserTempXToken(userTOKEN, fields, expiration, appName)
	if err == nil {
		t.Fatalf("Using unknown app, should fail.")
	}
	if userXToken != "" {
		t.Fatalf("Should fail to generate user token")
	}
	_, err = db.deleteUserRecord(userTOKEN)
	if err != nil {
		t.Fatalf("Failed to delete user: %s", err)
	}
}
