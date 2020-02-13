package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func Test_UUID(t *testing.T) {
	for id := 1; id < 11; id++ {
		recordUUID, err := uuid.GenerateUUID()
		t.Logf("Checking[%d]: %s\n", id, recordUUID)
		if err != nil {
			t.Fatalf("Failed to generate UUID %s: %s ", recordUUID, err)
		} else if isValidUUID(recordUUID) == false {
			t.Fatalf("Failed to validate UUID: %s ", recordUUID)
		}
	}
}

func Test_AppNames(t *testing.T) {
	goodApps := []string{"penn", "teller", "a123"}
	for _, value := range goodApps {
		if isValidApp(value) == false {
			t.Fatalf("Failed to validate good app name: %s ", value)
		}
	}
	badApps := []string{"P1", "1as", "_a", "a_a", "a.a", "a a"}
	for _, value := range badApps {
		if isValidApp(value) == true {
			t.Fatalf("Failed to validate bad app name: %s ", value)
		}
	}

}

func Test_stringPatternMatch(t *testing.T) {
	goodJsons := []map[string]interface{}{
		{"pattern": "*", "name": "tom", "result": true},
		{"pattern": "aa", "name": "tom", "result": false},
		{"pattern": "", "name": "aa", "result": false},
		{"pattern": "test*", "name": "123testabc", "result": false},
		{"pattern": "test*", "name": "testabc", "result": true},
		{"pattern": "*test*", "name": "test1", "result": true},
		{"pattern": "*test", "name": "123testabc", "result": false},
		{"pattern": "*test", "name": "123test", "result": true},
	}
	for _, value := range goodJsons {
		if stringPatternMatch(value["pattern"].(string), value["name"].(string)) != value["result"].(bool) {
			t.Fatalf("Failed in %s match %s\n", value["pattern"].(string), value["name"].(string))
		}
	}
}

func Test_getJSONPost(t *testing.T) {
	goodJsons := []string{
		`{"login":"abc","name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":"1234","name": "tom", "pass": "mylittlepony", "admin": true}`,
	}
	for _, value := range goodJsons {
		request := httptest.NewRequest("POST", "/user", strings.NewReader(value))
		request.Header.Set("Content-Type", "application/json")
		result, err := getJSONPost(request, "IL")
		if err != nil {
			t.Fatalf("Failed to parse json: %s, err: %s\n", value, err)
		}
		if len(result.loginIdx) == 0 {
			t.Fatalf("Failed to parse login index from json: %s ", value)
		}
	}

	badJsons := []string{
		`{"login":true,"name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":1,"name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":null,"name": "tom", "pass": "mylittlepony", "admin": true}`,
	}
	for _, value := range badJsons {
		request := httptest.NewRequest("POST", "/user", strings.NewReader(value))
		request.Header.Set("Content-Type", "application/json")
		result, err := getJSONPost(request, "IL")
		if err != nil {
			t.Fatalf("Failed to parse json: %s, err: %s\n", value, err)
		}
		if len(result.loginIdx) != 0 {
			t.Fatalf("Failed to parse login index from json: %s ", value)
		}
	}
}

func Test_SMS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %s\n", string(bodyBytes))
		if string(bodyBytes) != "Body=Data+Bunker+code+1234&From=from1234&To=4444" {
			t.Fatalf("bad request: %s", string(bodyBytes))
		}
	}))
	// Close the server when test finishes
	defer server.Close()
	client := server.Client()
	domain := server.URL
	var cfg Config
	cfg.Sms.TwilioToken = "ttoken"
	cfg.Sms.TwilioAccount = "taccount"
	cfg.Sms.TwilioFrom = "from1234"
	sendCodeByPhoneDo(domain, client, 1234, "4444", cfg)
}

func TestNotifyConsentChange(t *testing.T) {
	q := make(chan string)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %s\n", string(bodyBytes))
		if string(bodyBytes) != `{"action":"consentchange","address":"user3@user3.com","brief":"brief","mode":"email","status":"no"}` {
			q <- fmt.Sprintf("bad request in notifyConsentChange: %s", string(bodyBytes))
		} else {
			q <- "ok"
		}
	}))
	// Close the server when test finishes
	defer server.Close()
	notifyConsentChange(server.URL, "brief", "no", "email", "user3@user3.com")
	response := <-q
	if response != "ok" {
		t.Fatal(response)
	}
}

func TestNotifyProfileNew(t *testing.T) {
	q := make(chan string)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %s\n", string(bodyBytes))
		if string(bodyBytes) != `{"action":"profilenew","address":"user3@user3.com","mode":"email","profile":{"name":"alex"}}` {
			q <- fmt.Sprintf("bad request in notifyConsentChange: %s", string(bodyBytes))
		} else {
			q <- "ok"
		}
	}))
	// Close the server when test finishes
	defer server.Close()
	profile := []byte(`{"name":"alex"}`)
	notifyProfileNew(server.URL, profile, "email", "user3@user3.com")
	response := <-q
	if response != "ok" {
		t.Fatal(response)
	}
}

func TestNotifyForgetMe(t *testing.T) {
	q := make(chan string)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %s\n", string(bodyBytes))
		if string(bodyBytes) != `{"action":"forgetme","address":"user3@user3.com","mode":"email","profile":{"name":"alex"}}` {
			q <- fmt.Sprintf("bad request in notifyConsentChange: %s", string(bodyBytes))
		} else {
			q <- "ok"
		}
	}))
	// Close the server when test finishes
	defer server.Close()
	profile := []byte(`{"name":"alex"}`)
	notifyForgetMe(server.URL, profile, "email", "user3@user3.com")
	response := <-q
	if response != "ok" {
		t.Fatal(response)
	}
}

func TestNotifyProfileChange(t *testing.T) {
	q := make(chan string)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %s\n", string(bodyBytes))
		if string(bodyBytes) != `{"action":"profilechange","address":"user3@user3.com","mode":"email","old":{"name":"alex2"},"profile":{"name":"alex3"}}` {
			q <- fmt.Sprintf("bad request in notifyConsentChange: %s", string(bodyBytes))
		} else {
			q <- "ok"
		}
	}))
	// Close the server when test finishes
	defer server.Close()
	profile := []byte(`{"name":"alex2"}`)
	profile2 := []byte(`{"name":"alex3"}`)
	notifyProfileChange(server.URL, profile, profile2, "email", "user3@user3.com")
	response := <-q
	if response != "ok" {
		t.Fatal(response)
	}
}
