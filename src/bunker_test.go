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
	"github.com/securitybunker/databunker/src/storage"
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
	fmt.Printf("[%d] %s%s\n", rr.Code, request.Host, request.URL.Path)
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
	fmt.Printf("[%d] %s%s\n", rr.Code, request.Host, request.URL.Path)
	fmt.Printf("Response: %s\n", rr.Body.Bytes())
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
	fmt.Printf("[%d] %s%s\n", rr.Code, request.Host, request.URL.Path)
	fmt.Printf("Response: %s\n", rr.Body.Bytes())
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

func init() {
	fmt.Printf("**INIT*TEST*CODE***\n")
	testDBFile := storage.CreateTestDB()
	db, myRootToken, err := setupDB(&testDBFile, nil, "")
	if err != nil {
		//log.Panic("error %s", err.Error())
		fmt.Printf("error %s", err.Error())
	}
	rootToken = myRootToken
	var cfg Config
	cfile := "../databunker.yaml"
	err = readConfFile(&cfg, &cfile)
	cfg.SelfService.AppRecordChange = []string{"testapp", "super"}
	if err != nil {
		cfg.SelfService.ForgetMe = false
		cfg.SelfService.UserRecordChange = true
		cfg.Generic.CreateUserWithoutAccessToken = true
		//cfg.Generic.UseSeparateAppTables = true
		cfg.Policy.MaxUserRetentionPeriod = "1m"
		cfg.Policy.MaxAuditRetentionPeriod = "12m"
		cfg.Policy.MaxSessionRetentionPeriod = "1h"
		cfg.Policy.MaxShareableRecordRetentionPeriod = "1m"
	}
	e = mainEnv{db, cfg, make(chan struct{})}
	rootToken2, err := e.db.getRootXtoken()
	if err != nil {
		fmt.Printf("Failed to retrieve root token: %s\n", err)
	}
	fmt.Printf("Hashed root token: %s\n", rootToken2)
	router = e.setupRouter()
	router = e.setupConfRouter(router)
	//test1 := &testEnv{e, rootToken, router}
	e.dbCleanupDo()
	fmt.Printf("**INIT*DONE***\n")
}

func TestBackupOK(t *testing.T) {
	fmt.Printf("root token: %s\n", rootToken)
	raw, err := helpBackupRequest(rootToken)
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("failed to backup db %s", err.Error())
	}
	if strings.Contains(string(raw), "CREATE TABLE") == false {
		t.Fatalf("Backup failed\n")
	}
}

func TestMetrics(t *testing.T) {
	raw, err := helpMetricsRequest(rootToken)
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("failed to get metrics %s", err.Error())
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
		//if err != nil {
		//	log.Fatalf("failed to get page %s", err.Error())
		//}
		if strings.Contains(string(raw), pattern) == false {
			t.Fatalf("pattern detection failed\n")
		}
	}
}

func TestConfigurationOK(t *testing.T) {
	fmt.Printf("root token: %s\n", rootToken)
	raw, err := helpConfigurationDump(rootToken)
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("failed to fetch configuration: %s", err.Error())
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
