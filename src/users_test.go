package databunker

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func helpCreateUser(userJSON string) (map[string]interface{}, error) {
	request := httptest.NewRequest("POST", "http://localhost:3000/v1/user", strings.NewReader(userJSON))
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", rootToken)
	fmt.Printf("**** Using root token: %s\n", rootToken)
	router.ServeHTTP(rr, request)
	/*
		if status := rr.Code; status != http.StatusOK {
			err := errors.New("Wrong status")
			return nil, err
		}
	*/
	/*
		resp := rr.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			t.Fatalf("Status code: %d", resp.StatusCode)
		}
		t.Log(resp.Header.Get("Content-Type"))
		t.Log(string(body))
	*/

	var raw map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUser(index string, indexValue string) (map[string]interface{}, error) {
	request := httptest.NewRequest("GET", "http://localhost:3000/v1/user/"+index+"/"+indexValue, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpDeleteUser(index string, indexValue string) (map[string]interface{}, error) {
	request := httptest.NewRequest("DELETE", "http://localhost:3000/v1/user/"+index+"/"+indexValue, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserAuditEvents(userTOKEN string) (map[string]interface{}, error) {
	request := httptest.NewRequest("GET", "http://localhost:3000/v1/audit/list/"+userTOKEN, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func helpGetUserAuditEvent(atoken string) (map[string]interface{}, error) {
	url := "http://localhost:3000/v1/audit/get/" + atoken
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("X-Bunker-Token", rootToken)

	router.ServeHTTP(rr, request)
	var raw map[string]interface{}
	fmt.Printf("Got: %s\n", rr.Body.Bytes())
	err := json.Unmarshal(rr.Body.Bytes(), &raw)
	return raw, err
}

func TestPOSTCreateUser(t *testing.T) {

	userJSON := `{"login":"user1","name":"tom","pass":"mylittlepony","k1":[1,10,20],"k2":{"f1":"t1","f3":{"a":"b"}}}`

	raw, err := helpCreateUser(userJSON)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	var userTOKEN string
	if status, ok := raw["status"]; ok {
		if status == "error" {
			if strings.HasPrefix(raw["message"].(string), "duplicate") {
				//_, userUUID, _ = e.db.getUserIndex("user1", "login")
				//fmt.Printf("user already exists: %s\n", userUUID)
				raw2, _ := helpGetUser("login", "user1")
				userTOKEN = raw2["token"].(string)
			} else {
				t.Fatalf("Failed to create user: %s\n", raw["message"])
				return
			}
		} else if status == "ok" {
			userTOKEN = raw["token"].(string)
		}
	}
	if len(userTOKEN) == 0 {
		t.Fatalf("Failed to parse userTOKEN")
	}
	raw2, err := helpGetUserAuditEvents(userTOKEN)
	if raw2["status"].(string) != "ok" {
		t.Fatalf("Failed to get audit event/s\n")
	}
	if raw2["total"].(float64) != 1 {
		t.Fatalf("Wrong number of audit event/s\n")
	}
	records := raw2["rows"].([]interface{})
	records0 := records[0].(map[string]interface{})
	atoken := records0["atoken"].(string)
	if len(atoken) == 0 {
		t.Fatalf("Failed to extract atoken\n")
	}
	fmt.Printf("Audit record: %s\n", atoken)
	raw3, _ := helpGetUserAuditEvent(atoken)
	if raw3["status"].(string) != "ok" {
		t.Fatalf("Failed to get specific audit event\n")
	}
	helpDeleteUser("login", "user1")
	raw4, _ := helpGetUser("login", "user1")
	//userTOKEN = raw3["token"].(string)
	//fmt.Printf("status: %s", raw3["status"])
	if strings.Contains(raw4["message"].(string), "not found") == false {
		t.Fatalf("Failed to delete user, got message: %s", raw2["message"].(string))
	}
}
