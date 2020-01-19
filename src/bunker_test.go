package databunker

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
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

func helpBackupRequest(token string) ([]byte, error) {
	url := "http://localhost:3000/v1/sys/backup"
	request := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bunker-Token", token)

	router.ServeHTTP(rr, request)
	if rr.Code != 200 {
		return nil, fmt.Errorf("wrong status: %d", rr.Code)
	}
	//fmt.Printf("Got: %s\n", rr.Body.Bytes())
	return rr.Body.Bytes(), nil
}

func init() {
	fmt.Printf("**INIT*BEFORE***\n")
	masterKey, _ := hex.DecodeString("71c65924336c5e6f41129b6f0540ad03d2a8bf7e9b10db72")
	testDBFile := "/tmp/test.sqlite3"
	os.Remove(testDBFile)
	db, _ := newDB(masterKey, &testDBFile)
	err := db.initDB()
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("db init error %s", err.Error())
	}
	db.initUserApps()
	var cfg Config
	e := mainEnv{db, cfg, make(chan struct{})}
	rootToken, err = db.createRootXtoken()
	if err != nil {
		//log.Panic("error %s", err.Error())
		fmt.Printf("error %s", err.Error())
	}
	fmt.Printf("Root token: %s\n", rootToken)
	rootToken2, err := e.db.getRootXtoken()
	if err != nil {
		fmt.Printf("Failed to retreave root token: %s\n", err)
	}
	fmt.Printf("Hashed root token: %s\n", rootToken2)
	router = e.setupRouter()
	//test1 := &testEnv{e, rootToken, router}
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

func TestBackupError(t *testing.T) {
	token, _ := uuid.GenerateUUID()
	_, err := helpBackupRequest(token)
	if err == nil {
		log.Fatalf("This test should faile")
	}
}
