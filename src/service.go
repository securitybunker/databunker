package main

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
)

func loadService() {
	initPtr := flag.Bool("init", false, "Generate master key and init database")
	demoPtr := flag.Bool("demoinit", false, "Generate master key with a DEMO root access token")
	startPtr := flag.Bool("start", false, "Start databunker service. Provide additional --masterkey value or set it up using evironment variable: DATABUNKER_MASTERKEY")
	masterKeyPtr := flag.String("masterkey", "", "Specify master key - main database encryption key")
	dbPtr := flag.String("db", "databunker", "Specify database name/file")
	confPtr := flag.String("conf", "", "Configuration file name to use")
	rootTokenKeyPtr := flag.String("roottoken", "", "Specify custom root token to use during database init. It must be in UUID format.")
	versionPtr := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *versionPtr {
		log.Printf("Databunker version: %s\n", version)
		os.Exit(0)
	}
	log.Printf("Databunker version: %s\n", version)
	var cfg Config
	ReadEnv(&cfg)
	ReadConfFile(&cfg, confPtr)

	customRootToken := ""
	if *demoPtr {
		customRootToken = "DEMO"
	} else {
		customRootToken = utils.GetArgEnvFileVariable("DATABUNKER_ROOTTOKEN", rootTokenKeyPtr)
	}
	if *initPtr || *demoPtr {
		if storage.DBExists(dbPtr) == true {
			log.Println("Database is alredy initialized.")
		} else {
			db, _, _ := setupDB(dbPtr, masterKeyPtr, customRootToken)
			db.store.CloseDB()
		}
		os.Exit(0)
	}
	dbExists := storage.DBExists(dbPtr)
	for numAttempts := 60; dbExists == false && numAttempts > 0; numAttempts-- {
		time.Sleep(1 * time.Second)
		log.Printf("Trying to open db [%d]\n", 61-numAttempts)
		dbExists = storage.DBExists(dbPtr)
	}
	if dbExists == false {
		log.Println("Database is not initialized")
		log.Println(`Run "databunker -init" for the first time to generate keys and init database.`)
		os.Exit(0)
	}
	masterKeyStr := utils.GetArgEnvFileVariable("DATABUNKER_MASTERKEY", masterKeyPtr)
	if *startPtr == false {
		log.Println(`'databunker -start' command is missing.`)
		os.Exit(0)
	}
	if len(masterKeyStr) == 0 {
		log.Println(`ENV['DATABUNKER_MASTERKEY'], ENV['DATABUNKER_MASTERKEY_FILE'], or 'databunker -masterkey value' must be provided.`)
		os.Exit(0)
	}
	err := loadUserSchema(cfg, confPtr)
	if err != nil {
		log.Printf("Failed to load user schema: %s\n", err)
		os.Exit(0)
	}
	masterKey, masterKeyErr := decodeMasterkey(masterKeyStr)
	if masterKeyErr != nil {
		log.Printf("Error: %s", masterKeyErr)
		os.Exit(0)
	}
	store, err := storage.OpenDB(dbPtr)
	if err != nil {
		log.Printf("Filed to open db: %s", err)
		os.Exit(0)
	}
	hash := md5.Sum(masterKey)
	db := &dbcon{store, masterKey, hash[:]}
	e := mainEnv{db, cfg, make(chan struct{})}
	e.dbCleanup()
	initGeoIP()
	initCaptcha(hash)
	router := e.setupRouter()
	router = e.setupConfRouter(router)
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			//tls.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			//tls.TLS_DHE_RSA_WITH_AES_256_CCM_8,
			//tls.TLS_DHE_RSA_WITH_AES_256_CCM,
			//tls.TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384,
			//tls.TLS_DHE_RSA_WITH_ARIA_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
	listener := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{Addr: listener, Handler: e.reqMiddleware(router), TLSConfig: tlsConfig}

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// Waiting for SIGINT (pkill -2)
	go func() {
		<-stop
		log.Println("Closing app...")
		close(e.stopChan)
		time.Sleep(1 * time.Second)
		srv.Shutdown(context.TODO())
		db.store.CloseDB()
	}()

	if _, err := os.Stat(cfg.Ssl.SslCertificate); !os.IsNotExist(err) {
		log.Printf("Open HTTPS listener %s\n", listener)
		err := srv.ListenAndServeTLS(cfg.Ssl.SslCertificate, cfg.Ssl.SslCertificateKey)
		if err != nil {
			log.Printf("ListenAndServeSSL: %s\n", err)
		}
	} else {
		log.Printf("Open HTTP listener %s\n", listener)
		err := srv.ListenAndServe()
		if err != nil {
			log.Printf("ListenAndServe(): %s\n", err)
		}
	}
}

func decodeMasterkey(masterKeyStr string) ([]byte, error) {
	if len(masterKeyStr) == 0 {
		return nil, errors.New("Master key environment variable/parameter is missing")
	}
	if len(masterKeyStr) != 48 {
		return nil, errors.New("Master key length is wrong")
	}
	if utils.CheckValidHex(masterKeyStr) == false {
		return nil, errors.New("Master key is not valid hex string")
	}
	masterKey, err := hex.DecodeString(masterKeyStr)
	if err != nil {
		return nil, errors.New("Failed to decode master key")
	}
	return masterKey, nil
}

func setupDB(dbPtr *string, masterKeyPtr *string, customRootToken string) (*dbcon, string, error) {
	log.Println("Databunker init")
	var masterKey []byte
	var err error
	masterKeyString := utils.GetArgEnvFileVariable("DATABUNKER_MASTERKEY", masterKeyPtr)
	if len(masterKeyString) > 0 {
		masterKey, err = decodeMasterkey(masterKeyString)
		if err != nil {
			log.Printf("Failed to parse master key: %s\n", err)
			os.Exit(0)
		}
		log.Println("Master key: ****")
	} else {
		masterKey, err = utils.GenerateMasterKey()
		if err != nil {
			log.Printf("Failed to generate master key: %s", err)
			os.Exit(0)
		}
		log.Printf("Master key: %x\n", masterKey)
	}
	hash := md5.Sum(masterKey)
	log.Println("Init database")
	store, err := storage.InitDB(dbPtr)
	for numAttempts := 60; err != nil && numAttempts > 0; numAttempts-- {
		time.Sleep(1 * time.Second)
		log.Printf("Trying to init db: %d\n", 61-numAttempts)
		store, err = storage.InitDB(dbPtr)
	}
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("Databunker failed to init database, error %s\n\n", err.Error())
		os.Exit(0)
	}
	db := &dbcon{store, masterKey, hash[:]}
	rootToken, err := db.createRootXtoken(customRootToken)
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Printf("Failed to init root token: %s", err.Error())
		os.Exit(0)
	}
	log.Println("Creating default legal basis records")
	db.createLegalBasis("core-send-email-on-login", "", "login", "Send email on login",
		"Confirm to allow sending access code using 3rd party email gateway", "consent",
		"This consent is required to give you our service.", "active", true, true)
	db.createLegalBasis("core-send-sms-on-login", "", "login", "Send SMS on login",
		"Confirm to allow sending access code using 3rd party SMS gateway", "consent",
		"This consent is required to give you our service.", "active", true, true)
	if len(customRootToken) > 0 && customRootToken != "DEMO" {
		log.Println("API Root token: ****")
	} else {
		log.Printf("API Root token: %s\n", rootToken)
	}
	return db, rootToken, err
}
