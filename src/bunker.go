// Package main - Personal Identifiable Information (PII) database.
// For more info check https://databunker.org/
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/securitybunker/databunker/src/autocontext"
	"github.com/securitybunker/databunker/src/storage"
	yaml "gopkg.in/yaml.v2"
)

type dbcon struct {
	store     storage.BackendDB
	masterKey []byte
	hash      []byte
}

// Config is u	sed to store application configuration
type Config struct {
	Generic struct {
		CreateUserWithoutAccessToken bool   `yaml:"create_user_without_access_token" default:"false"`
		UseSeparateAppTables         bool   `yaml:"use_separate_app_tables" default:"false"`
		UserRecordSchema             string `yaml:"user_record_schema"`
		AdminEmail                   string `yaml:"admin_email" envconfig:"ADMIN_EMAIL"`
		ListUsers                    bool   `yaml:"list_users" default:"false"`
	}
	SelfService struct {
		ForgetMe         bool     `yaml:"forget_me" default:"false"`
		UserRecordChange bool     `yaml:"user_record_change" default:"false"`
		AppRecordChange  []string `yaml:"app_record_change"`
	}
	Notification struct {
		NotificationURL string `yaml:"notification_url"`
		MagicSyncURL    string `yaml:"magic_sync_url"`
		MagicSyncToken  string `yaml:"magic_sync_token"`
	}
	Policy struct {
		MaxUserRetentionPeriod            string `yaml:"max_user_retention_period" default:"1m"`
		MaxAuditRetentionPeriod           string `yaml:"max_audit_retention_period" default:"12m"`
		MaxSessionRetentionPeriod         string `yaml:"max_session_retention_period" default:"1h"`
		MaxShareableRecordRetentionPeriod string `yaml:"max_shareable_record_retention_period" default:"1m"`
	}
	Ssl struct {
		SslCertificate    string `yaml:"ssl_certificate" envconfig:"SSL_CERTIFICATE"`
		SslCertificateKey string `yaml:"ssl_certificate_key" envconfig:"SSL_CERTIFICATE_KEY"`
	}
	Sms struct {
		URL            string `yaml:"url"`
		From           string `yaml:"from"`
		Body           string `yaml:"body"`
		Token          string `yaml:"token"`
		Method         string `yaml:"method"`
		BasicAuth      string `yaml:"basic_auth"`
		ContentType    string `yaml:"content_type"`
		CustomHeader   string `yaml:"custom_header"`
		DefaultCountry string `yaml:"default_country"`
	}
	Server struct {
		Port string `yaml:"port" envconfig:"BUNKER_PORT"`
		Host string `yaml:"host" envconfig:"BUNKER_HOST"`
	} `yaml:"server"`
	SMTP struct {
		Server string `yaml:"server" envconfig:"SMTP_SERVER"`
		Port   string `yaml:"port" envconfig:"SMTP_PORT"`
		User   string `yaml:"user" envconfig:"SMTP_USER"`
		Pass   string `yaml:"pass" envconfig:"SMTP_PASS"`
		Sender string `yaml:"sender" envconfig:"SMTP_SENDER"`
	} `yaml:"smtp"`
	UI struct {
		LogoLink           string `yaml:"logo_link"`
		CompanyTitle       string `yaml:"company_title"`
		CompanyVAT         string `yaml:"company_vat"`
		CompanyCity        string `yaml:"company_city"`
		CompanyLink        string `yaml:"company_link"`
		CompanyCountry     string `yaml:"company_country"`
		CompanyAddress     string `yaml:"company_address"`
		TermOfServiceTitle string `yaml:"term_of_service_title"`
		TermOfServiceLink  string `yaml:"term_of_service_link"`
		PrivacyPolicyTitle string `yaml:"privacy_policy_title"`
		PrivacyPolicyLink  string `yaml:"privacy_policy_link"`
		CustomCSSLink      string `yaml:"custom_css_link"`
		MagicLookup        bool   `yaml:"magic_lookup"`
	} `yaml:"ui"`
}

// mainEnv struct stores global structures
type mainEnv struct {
	db       *dbcon
	conf     Config
	stopChan chan struct{}
}

// userJSON used to parse user POST
type userJSON struct {
	jsonData  []byte
	loginIdx  string
	emailIdx  string
	phoneIdx  string
	customIdx string
	token     string
}

type tokenAuthResult struct {
	ttype string
	name  string
	token string
}

type checkRecordResult struct {
	name    string
	token   string
	fields  string
	appName string
	session string
}

func prometheusHandler() http.Handler {
	handlerOptions := promhttp.HandlerOpts{
		ErrorHandling:      promhttp.ContinueOnError,
		DisableCompression: true,
	}
	promHandler := promhttp.HandlerFor(prometheus.DefaultGatherer, handlerOptions)
	return promHandler
}

// metrics API call
func (e mainEnv) metrics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Printf("/metrics\n")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	//log.Fprintf(w, `{"status":"ok","apps":%q}`, result)
	//log.Fprintf(w, "hello")
	//promhttp.Handler().ServeHTTP(w, r)
	prometheusHandler().ServeHTTP(w, r)
}

// backupDB API call.
func (e mainEnv) backupDB(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	w.WriteHeader(200)
	e.db.store.BackupDB(w)
}

func (e mainEnv) checkStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := e.db.store.Ping()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error"))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}
}

// setupRouter() setup HTTP Router object.
func (e mainEnv) setupRouter() *httprouter.Router {
	box := packr.NewBox("../ui")

	router := httprouter.New()

	router.GET("/v1/status", e.checkStatus)
	router.GET("/v1/status/", e.checkStatus)
	router.GET("/status", e.checkStatus)
	router.GET("/status/", e.checkStatus)

	router.GET("/v1/sys/backup", e.backupDB)

	router.POST("/v1/user", e.userCreate)
	router.POST("/v1/users", e.userList)
	router.GET("/v1/user/:mode/:identity", e.userGet)
	router.DELETE("/v1/user/:mode/:identity", e.userDelete)
	router.PUT("/v1/user/:mode/:identity", e.userChange)

	router.GET("/v1/prelogin/:mode/:identity/:code/:captcha", e.userPrelogin)
	router.GET("/v1/login/:mode/:identity/:tmp", e.userLogin)

	router.GET("/v1/exp/retain/:exptoken", e.expRetainData)
	router.GET("/v1/exp/delete/:exptoken", e.expDeleteData)
	router.GET("/v1/exp/status/:mode/:identity", e.expGetStatus)
	router.POST("/v1/exp/start/:mode/:identity", e.expStart)
	router.DELETE("/v1/exp/cancel/:mode/:identity", e.expCancel)

	router.POST("/v1/sharedrecord/token/:token", e.newSharedRecord)
	router.GET("/v1/get/:record", e.getRecord)

	router.GET("/v1/request/:request", e.getUserRequest)
	router.POST("/v1/request/:request", e.approveUserRequest)
	router.DELETE("/v1/request/:request", e.cancelUserRequest)
	router.GET("/v1/requests/:mode/:identity", e.getCustomUserRequests)
	router.GET("/v1/requests", e.getUserRequests)

	router.GET("/v1/pactivity", e.pactivityList)
	router.POST("/v1/pactivity/:activity", e.pactivityCreate)
	router.DELETE("/v1/pactivity/:activity", e.pactivityDelete)
	router.POST("/v1/pactivity/:activity/:brief", e.pactivityLink)
	router.DELETE("/v1/pactivity/:activity/:brief", e.pactivityUnlink)

	router.GET("/v1/lbasis", e.listLegalBasisRecords)
	router.POST("/v1/lbasis/:brief", e.createLegalBasis)
	router.DELETE("/v1/lbasis/:brief", e.deleteLegalBasis)

	router.GET("/v1/agreement/:brief/:mode/:identity", e.getUserAgreement)
	router.POST("/v1/agreement/:brief/:mode/:identity", e.agreementAccept)
	router.DELETE("/v1/agreement/:brief", e.agreementRevokeAll)
	router.DELETE("/v1/agreement/:brief/:mode/:identity", e.agreementWithdraw)
	router.GET("/v1/agreements/:mode/:identity", e.getUserAgreements)

	//router.GET("/v1/consent/:mode/:identity", e.consentAllUserRecords)
	//router.GET("/v1/consent/:mode/:identity/:brief", e.consentUserRecord)

	router.POST("/v1/userapp/token/:token/:appname", e.userappNew)
	router.GET("/v1/userapp/token/:token/:appname", e.userappGet)
	router.PUT("/v1/userapp/token/:token/:appname", e.userappChange)
	router.DELETE("/v1/userapp/token/:token/:appname", e.userappDelete)
	router.GET("/v1/userapp/token/:token", e.userappList)
	router.GET("/v1/userapps", e.appList)

	router.GET("/v1/session/:session", e.getSession)
	router.POST("/v1/session/:session", e.createSession)
	router.DELETE("/v1/session/:session", e.deleteSession)
	//router.POST("/v1/sessions/:mode/:identity", e.newUserSession)
	router.GET("/v1/sessions/:mode/:identity", e.getUserSessions)

	router.GET("/v1/metrics", e.metrics)

	router.GET("/v1/audit/admin", e.getAdminAuditEvents)
	router.GET("/v1/audit/list/:token", e.getAuditEvents)
	router.GET("/v1/audit/get/:atoken", e.getAuditEvent)

	router.GET("/v1/captcha/:code", e.showCaptcha)

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		data, err := box.Find("index.html")
		if err != nil {
			//log.Panic("error %s", err.Error())
			log.Printf("error: %s\n", err.Error())
			w.WriteHeader(404)
		} else {
			captcha, err := generateCaptcha()
			if err != nil {
				w.WriteHeader(501)
			} else {
				data2 := bytes.ReplaceAll(data, []byte("%CAPTCHAURL%"), []byte(captcha))
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(200)
				w.Write(data2)
			}
		}
	})
	router.GET("/site/*filepath", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fname := r.URL.Path
		if fname == "/site/" {
			fname = "/site/index.html"
		}
		data, err := box.Find(fname)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("url not found"))
		} else {
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			} else if strings.HasSuffix(r.URL.Path, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			} else if strings.HasSuffix(r.URL.Path, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(r.URL.Path, ".woff2") {
				w.Header().Set("Content-Type", "font/woff2")
			} else if strings.HasSuffix(r.URL.Path, ".html") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			}
			// next step: https://www.sanarias.com/blog/115LearningHTTPcachinginGo
			w.Header().Set("Cache-Control", "public, max-age=7776000")
			w.WriteHeader(200)
			if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".html") {
				if bytes.Contains(data, []byte("%IPCOUNTRY%")) {
					ipCountry := getCountry(r)
					data2 := bytes.ReplaceAll(data, []byte("%IPCOUNTRY%"), []byte(ipCountry))
					w.Write([]byte(data2))
					return
				}
			}
			w.Write([]byte(data))
		}
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":"error", "message":"url not found"}`))
	})
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//if r.Header.Get("Access-Control-Request-Method") != "" {
		// Set CORS headers
		header := w.Header()
		header.Set("Access-Control-Allow-Methods", "POST, PUT, DELETE")
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Headers", "Accept,  Content-Type, Content-Length, Accept-Encoding, X-Bunker-Token")
		//}
		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})
	return router
}

// readConfFile() read configuration file.
func readConfFile(cfg *Config, filepath *string) error {
	confFile := "databunker.yaml"
	if filepath != nil {
		if len(*filepath) > 0 {
			confFile = *filepath
		}
	}
	fmt.Printf("Databunker configuration file is: %s\n", confFile)
	f, err := os.Open(confFile)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}
	return nil
}

// readEnv() process environment variables.
func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	return err
}

// dbCleanup() is used to run cron jobs.
func (e mainEnv) dbCleanupDo() {
	log.Printf("db cleanup timeout\n")
	exp, _ := parseExpiration0(e.conf.Policy.MaxAuditRetentionPeriod)
	if exp > 0 {
		e.db.store.DeleteExpired0(storage.TblName.Audit, exp)
	}
	notifyURL := e.conf.Notification.NotificationURL
	e.db.expireAgreementRecords(notifyURL)
	e.expUsers()
}

func (e mainEnv) dbCleanup() {
	ticker := time.NewTicker(time.Duration(10) * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				e.dbCleanupDo()
			case <-e.stopChan:
				log.Printf("db cleanup closed\n")
				ticker.Stop()
				return
			}
		}
	}()
}

// CustomResponseWriter struct is a custom wrapper for ResponseWriter
type CustomResponseWriter struct {
	w    http.ResponseWriter
	gz   io.Writer
	Code int
}

// NewCustomResponseWriter function returns CustomResponseWriter object
func NewCustomResponseWriter(ww http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		w:    ww,
		gz:   nil,
		Code: 0,
	}
}

// Header function returns HTTP Header object
func (w *CustomResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.w.Write(b)
}

// WriteHeader function writes header back to original ResponseWriter
func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
	w.w.WriteHeader(statusCode)
}

var statusCounter = 0
var statusErrorCounter = 0

func reqMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("Set host %s\n", r.Host)
		autocontext.Set(r, "host", r.Host)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy", "default-src 'self' http: https: data: blob: 'unsafe-inline'")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w2 := NewCustomResponseWriter(w)
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w2.Header().Set("Vary", "Accept-Encoding")
			w2.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			w2.gz = gz
			defer gz.Close()
		}
		handler.ServeHTTP(w2, r)
		autocontext.Clean(r)
		URL := r.URL.String()
		if r.Method == "GET" && (URL == "/status" || URL == "/status/" || URL == "/v1/status" || URL == "/v1/status/") {
			if w2.Code == 200 {
				if statusCounter < 2 {
					log.Printf("%d %s %s\n", w2.Code, r.Method, r.URL)
				} else if statusCounter == 2 {
					log.Printf("%d %s %s 'ignore subsequent /status requests'\n", w2.Code, r.Method, r.URL)
				}
				statusCounter = statusCounter + 1
			} else {
				if statusErrorCounter < 2 {
					log.Printf("%d %s %s\n", w2.Code, r.Method, r.URL)
				} else if statusErrorCounter == 2 {
					log.Printf("%d %s %s 'ignore subsequent errors in /status requests'\n", w2.Code, r.Method, r.URL)
				}
				statusErrorCounter = statusErrorCounter + 1
			}
		} else {
			statusCounter = 0
			statusErrorCounter = 0
			log.Printf("%d %s %s\n", w2.Code, r.Method, r.URL)
		}
	})
}

func setupDB(dbPtr *string, masterKeyPtr *string, customRootToken string) (*dbcon, string, error) {
	fmt.Printf("\nDatabunker init\n\n")
	var masterKey []byte
	var err error
	if variableProvided("DATABUNKER_MASTERKEY", masterKeyPtr) == true {
		masterKey, err = masterkeyGet(masterKeyPtr)
		if err != nil {
			fmt.Printf("Failed to parse master key: %s", err)
			os.Exit(0)
		}
		fmt.Printf("Master key: ****\n\n")
	} else {
		masterKey, err = generateMasterKey()
		if err != nil {
			fmt.Printf("Failed to generate master key: %s", err)
			os.Exit(0)
		}
		fmt.Printf("Master key: %x\n\n", masterKey)
	}
	hash := md5.Sum(masterKey)
	fmt.Printf("Init database\n\n")
	store, err := storage.InitDB(dbPtr)
	for numAttempts := 60; err != nil && numAttempts > 0; numAttempts-- {
		time.Sleep(1 * time.Second)
		fmt.Printf("Trying to init db: %d\n", 61-numAttempts)
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
		fmt.Printf("Failed to init root token: %s", err.Error())
		os.Exit(0)
	}
	log.Println("Creating default entities: core-send-email-on-login and core-send-sms-on-login")
	db.createLegalBasis("core-send-email-on-login", "", "login", "Send email on login",
		"Confirm to allow sending access code using 3rd party email gateway", "consent",
		"This consent is required to give you our service.", "active", true, true)
	db.createLegalBasis("core-send-sms-on-login", "", "login", "Send SMS on login",
		"Confirm to allow sending access code using 3rd party SMS gateway", "consent",
		"This consent is required to give you our service.", "active", true, true)
	if len(customRootToken) > 0 && customRootToken != "DEMO" {
		fmt.Printf("\nAPI Root token: ****\n\n")
	} else {
		fmt.Printf("\nAPI Root token: %s\n\n", rootToken)
	}
	return db, rootToken, err
}

func variableProvided(vname string, masterKeyPtr *string) bool {
	if masterKeyPtr != nil && len(*masterKeyPtr) > 0 {
		return true
	}
	if len(os.Getenv(vname)) > 0 {
		return true
	}
	if len(os.Getenv(vname+"_FILE")) > 0 {
		return true
	}
	return false
}

func masterkeyGet(masterKeyPtr *string) ([]byte, error) {
	masterKeyStr := ""
	if masterKeyPtr != nil && len(*masterKeyPtr) > 0 {
		masterKeyStr = *masterKeyPtr
	} else if len(os.Getenv("DATABUNKER_MASTERKEY")) > 0 {
		masterKeyStr = os.Getenv("DATABUNKER_MASTERKEY")
	} else if len(os.Getenv("DATABUNKER_MASTERKEY_FILE")) > 0 {
		content, err := ioutil.ReadFile(os.Getenv("DATABUNKER_MASTERKEY_FILE"))
		if err != nil {
			return nil, err
		}
		// Convert []byte to string
		masterKeyStr = strings.TrimSpace(string(content))
		// we will TRY to delete secret file when running inside container/kubernetes
		if isContainer() == true {
			os.Remove(os.Getenv("DATABUNKER_MASTERKEY_FILE"))
		}
	}
	if len(masterKeyStr) == 0 {
		return nil, errors.New("Master key environment variable/parameter is missing")
	}
	if len(masterKeyStr) != 48 {
		return nil, errors.New("Master key length is wrong")
	}
	if isValidHex(masterKeyStr) == false {
		return nil, errors.New("Master key is not valid hex string")
	}
	masterKey, err := hex.DecodeString(masterKeyStr)
	if err != nil {
		return nil, errors.New("Failed to decode master key")
	}
	return masterKey, nil
}

// main application function
func main() {
	rand.Seed(time.Now().UnixNano())
	lockMemory()
	initPtr := flag.Bool("init", false, "Generate master key and init database")
	demoPtr := flag.Bool("demoinit", false, "Generate master key with a DEMO root access token")
	startPtr := flag.Bool("start", false, "Start databunker service. Provide additional --masterkey value or set it up using evironment variable: DATABUNKER_MASTERKEY")
	masterKeyPtr := flag.String("masterkey", "", "Specify master key - main database encryption key")
	dbPtr := flag.String("db", "databunker", "Specify database name/file")
	confPtr := flag.String("conf", "", "Configuration file name to use")
	rootTokenKeyPtr := flag.String("roottoken", "", "Specify custom root token to use during database init. It must be in UUID format.")
	flag.Parse()

	var cfg Config
	readEnv(&cfg)
	readConfFile(&cfg, confPtr)

	customRootToken := ""
	if *demoPtr {
		customRootToken = "DEMO"
	} else if variableProvided("DATABUNKER_ROOTTOKEN", rootTokenKeyPtr) == true {
		if rootTokenKeyPtr != nil && len(*rootTokenKeyPtr) > 0 {
			customRootToken = *rootTokenKeyPtr
		} else {
			customRootToken = os.Getenv("DATABUNKER_ROOTTOKEN")
		}
	}
	if *initPtr || *demoPtr {
		if storage.DBExists(dbPtr) == true {
			fmt.Printf("\nDatabase is alredy initialized.\n\n")
		} else {
			db, _, _ := setupDB(dbPtr, masterKeyPtr, customRootToken)
			db.store.CloseDB()
		}
		os.Exit(0)
	}
	dbExists := storage.DBExists(dbPtr)
	for numAttempts := 60; dbExists == false && numAttempts > 0; numAttempts-- {
		time.Sleep(1 * time.Second)
		fmt.Printf("Trying to open db: %d\n", 61-numAttempts)
		dbExists = storage.DBExists(dbPtr)
	}
	if dbExists == false {
		fmt.Printf("\nDatabase is not initialized.\n\n")
		fmt.Println(`Run "databunker -init" for the first time to generate keys and init database.`)
		fmt.Println("")
		os.Exit(0)
	}
	if masterKeyPtr == nil && *startPtr == false {
		fmt.Println("")
		fmt.Println(`Run "databunker -start" will load DATABUNKER_MASTERKEY environment variable.`)
		fmt.Println(`For testing "databunker -masterkey MASTER_KEY_VALUE" can be used. Not recommended for production.`)
		fmt.Println("")
		os.Exit(0)
	}
	err := loadUserSchema(cfg, confPtr)
	if err != nil {
		fmt.Printf("Failed to load user schema: %s\n", err)
		os.Exit(0)
	}
	masterKey, masterKeyErr := masterkeyGet(masterKeyPtr)
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
	srv := &http.Server{Addr: cfg.Server.Host + ":" + cfg.Server.Port, Handler: reqMiddleware(router), TLSConfig: tlsConfig}

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
		log.Printf("Loading ssl\n")
		err := srv.ListenAndServeTLS(cfg.Ssl.SslCertificate, cfg.Ssl.SslCertificateKey)
		if err != nil {
			log.Printf("ListenAndServeSSL: %s\n", err)
		}
	} else {
		log.Println("Loading server")
		err := srv.ListenAndServe()
		if err != nil {
			log.Printf("ListenAndServe(): %s\n", err)
		}
	}
}
