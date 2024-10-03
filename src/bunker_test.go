package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/julienschmidt/httprouter"
)

var (
	e         mainEnv
	rootToken string
	router    *httprouter.Router
)

func helpServe0(request *http.Request) ([]byte, error) {
	request.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, request)
	log.Printf("[%d] %s%s\n", rr.Code, request.Host, request.URL.Path)
	if rr.Code != 200 {
		return rr.Body.Bytes(), fmt.Errorf("wrong status: %d", rr.Code)
	}
	//fmt.Printf("Response: %s\n", rr.Body.Bytes())
	return rr.Body.Bytes(), nil
}

func helpServe(request *http.Request) (map[string]interface{}, error) {
	request.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, request)
	log.Printf("[%d] %s%s %s\n", rr.Code, request.Host, request.URL.Path, rr.Body.Bytes())
	var raw map[string]interface{}
	if rr.Body.Bytes()[0] == '{' {
		json.Unmarshal(rr.Body.Bytes(), &raw)
	}
	if rr.Code != 200 {
		return raw, fmt.Errorf("wrong status: %d", rr.Code)
	}
	return raw, nil
}

func helpServe2(request *http.Request) (map[string]interface{}, error) {
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, request)
	log.Printf("[%d] %s%s %s\n", rr.Code, request.Host, request.URL.Path, rr.Body.Bytes())
	var raw map[string]interface{}
	if rr.Body.Bytes()[0] == '{' {
		json.Unmarshal(rr.Body.Bytes(), &raw)
	}
	if rr.Code != 200 {
		return raw, fmt.Errorf("wrong status: %d", rr.Code)
	}
	return raw, nil
}

func helpBackupRequest(token string) ([]byte, error) {
	url := "http://localhost:3000/v1/sys/backup"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", token)
	return helpServe0(request)
}

func helpMetricsRequest(token string) ([]byte, error) {
	url := "http://localhost:3000/v1/metrics"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", token)
	return helpServe0(request)
}

func helpConfigurationDump(token string) ([]byte, error) {
	url := "http://localhost:3000/v1/sys/configuration"
	request := httptest.NewRequest("GET", url, nil)
	request.Header.Set("X-Bunker-Token", token)
	return helpServe0(request)
}

/*
func TestBackupOK(t *testing.T) {
	log.Printf("root token: %s\n", rootToken)
	raw, err := helpBackupRequest(rootToken)
	if err != nil {
		log.Fatalf("Failed to backup db %s", err.Error())
	}
	if strings.Contains(string(raw), "CREATE TABLE") == false {
		t.Fatalf("Backup failed\n")
	}
}
*/

func TestMetrics(t *testing.T) {
	raw, err := helpMetricsRequest(rootToken)
	if err != nil {
		log.Fatalf("Failed to get metrics %s", err.Error())
	}
	if strings.Contains(string(raw), "go_memstats") == false {
		t.Fatalf("metrics failed\n")
	}
}

func TestAnonPage(t *testing.T) {
	goodJsons := []map[string]interface{}{
		{"url": "/", "pattern": "login"},
		{"url": "/site/", "pattern": "document.location"},
		{"url": "/site/site.js", "pattern": "dateFormat"},
		{"url": "/site/style.css", "pattern": "html"},
		{"url": "/site/user-profile.html", "pattern": "profile"},
		{"url": "/not-fund-page.html", "pattern": "not found"},
		{"url": "/site/not-fund-page.html", "pattern": "not found"},
	}
	for _, value := range goodJsons {
		url := "http://localhost:3000" + value["url"].(string)
		pattern := value["pattern"].(string)
		request := httptest.NewRequest("GET", url, nil)
		raw, _ := helpServe0(request)
		if strings.Contains(string(raw), pattern) == false {
			t.Fatalf("pattern detection failed\n")
		}
	}
}

func TestConfigurationOK(t *testing.T) {
	log.Printf("Generated root token: %s\n", rootToken)
	raw, err := helpConfigurationDump(rootToken)
	if err != nil {
		log.Fatalf("Failed to fetch configuration: %s", err.Error())
	}
	if strings.Contains(string(raw), "CreateUserWithoutAccessToken") == false {
		t.Fatalf("Configuration dump failed\n")
	}
}

func TestBackupError(t *testing.T) {
	token, _ := uuid.GenerateUUID()
	_, err := helpBackupRequest(token)
	if err == nil {
		log.Fatalf("This test should faile")
	}
}
