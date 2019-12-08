package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

/*
type testEnv struct {
	e         mainEnv
	rootToken string
	router    *httprouter.Router
}
*/

var (
	e         mainEnv
	rootToken string
	router    *httprouter.Router
)

func init() {
	fmt.Printf("***********BEFORE***\n")
	masterKey, _ := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	testFile := "/tmp/test"
	db, _ := newDB(masterKey, &testFile)
	var cfg Config
	e := mainEnv{db, cfg}
	db.initDB()
	var err error
	rootToken, err = db.createRootToken()
	if err != nil {
		//log.Panic("error %s", err.Error())
		fmt.Printf("error %s", err.Error())
	}
	fmt.Printf("Root token: %s\n", rootToken)
	rootToken, err = e.db.getRootToken()
	if err != nil {
		fmt.Printf("Failed to retreave root token: %s\n", err)
	}
	fmt.Printf("Loaded root token: %s\n", rootToken)
	router = e.setupRouter()
	//test1 := &testEnv{e, rootToken, router}
}

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

	helpDeleteUser("login", "user1")
	raw2, _ := helpGetUser("login", "user1")
	//userTOKEN = raw2["token"].(string)
	//fmt.Printf("status: %s", raw2["status"])
	if raw2["message"].(string) != "not found" {
		t.Fatalf("Failed to delete user, got message: %s", raw2["message"].(string))
	}
}
