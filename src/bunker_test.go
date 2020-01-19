package databunker

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	
	"github.com/julienschmidt/httprouter"
)

var (
	e         mainEnv
	rootToken string
	router    *httprouter.Router
)

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
