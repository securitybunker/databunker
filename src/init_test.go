package main

import (
	"log"

	"github.com/securitybunker/databunker/src/storage"
)

func init() {
	log.Printf("**INIT*TEST*CODE***\n")
	testDBFile := storage.CreateTestDB()
	db, myRootToken, err := setupDB(&testDBFile, nil, "")
	if err != nil {
		log.Printf("Init error %s", err.Error())
	}
	rootToken = myRootToken
	var cfg Config
	cfile := "../databunker.yaml"
	err = ReadConfFile(&cfg, &cfile)
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
		log.Printf("Failed to retrieve root token: %s\n", err)
	}
	log.Printf("Hashed root token: %s\n", rootToken2)
	router = e.setupRouter()
	router = e.setupConfRouter(router)
	//test1 := &testEnv{e, rootToken, router}
	e.dbCleanupDo()
	log.Printf("**INIT*DONE***\n")
}
