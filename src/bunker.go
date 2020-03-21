// Package main - Personal Identifiable Information (PII) database.
// For more info check https://paranoidguy.com
package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
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
	"github.com/paranoidguy/databunker/src/autocontext"
	"github.com/paranoidguy/databunker/src/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	yaml "gopkg.in/yaml.v2"
)

type dbcon struct {
	store     storage.DBStorage
	masterKey []byte
	hash      []byte
}

// Config is u	sed to store application configuration
type Config struct {
	Generic struct {
		CreateUserWithoutAccessToken bool   `yaml:"create_user_without_access_token"`
		AdminEmail                   string `yaml:"admin_email"`
	}
	SelfService struct {
		ForgetMe         bool     `yaml:"forget_me"`
		UserRecordChange bool     `yaml:"user_record_change"`
		AppRecordChange  []string `yaml:"app_record_change"`
		ConsentWithdraw  []string `yaml:"consent_withdraw"`
	}
	Notification struct {
		NotificationURL string `yaml:"notification_url"`
		MagicSyncURL    string `yaml:"magic_sync_url"`
		MagicSyncToken  string `yaml:"magic_sync_token"`
	}
	Policy struct {
		MaxAuditRetentionPeriod           string `yaml:"max_audit_retention_period"`
		MaxSessionRetentionPeriod         string `yaml:"max_session_retention_period"`
		MaxShareableRecordRetentionPeriod string `yaml:"max_shareable_record_retention_period"`
	}
	Ssl struct {
		SslCertificate    string `yaml:"ssl_certificate", envconfig:"SSL_CERTIFICATE"`
		SslCertificateKey string `yaml:"ssl_certificate_key", envconfig:"SSL_CERTIFICATE_KEY"`
	}
	Sms struct {
		DefaultCountry string `yaml:"default_country"`
		TwilioAccount  string `yaml:"twilio_account"`
		TwilioToken    string `yaml:"twilio_token"`
		TwilioFrom     string `yaml:"twilio_from"`
	}
	Server struct {
		Port string `yaml:"port", envconfig:"BUNKER_PORT"`
		Host string `yaml:"host", envconfig:"BUNKER_HOST"`
	} `yaml:"server"`
	SMTP struct {
		Server string `yaml:"server", envconfig:"SMTP_SERVER"`
		Port   string `yaml:"port", envconfig:"SMTP_PORT"`
		User   string `yaml:"user", envconfig:"SMTP_USER"`
		Pass   string `yaml:"pass", envconfig:"SMTP_PASS"`
		Sender string `yaml:"sender", envconfig:"SMTP_SENDER"`
	} `yaml:"smtp"`
	UI struct {
		LogoLink           string `yaml:"logo_link"`
		CompanyTitle       string `yaml:"company_title"`
		CompanyLink        string `yaml:"company_link"`
		TermOfServiceTitle string `yaml:"term_of_service_title"`
		TermOfServiceLink  string `yaml:"term_of_service_link"`
		PrivacyPolicyTitle string `yaml:"privacy_policy_title"`
		PrivacyPolicyLink  string `yaml:"privacy_policy_link"`
		CustomCSSFile      string `yaml:"custom_css_file"`
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
	jsonData []byte
	loginIdx string
	emailIdx string
	phoneIdx string
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
	fmt.Printf("/metrics\n")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	//fmt.Fprintf(w, `{"status":"ok","apps":%q}`, result)
	//fmt.Fprintf(w, "hello")
	//promhttp.Handler().ServeHTTP(w, r)
	prometheusHandler().ServeHTTP(w, r)
}

// configuration dump API call.
func (e mainEnv) configurationDump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	resultJSON, _ := json.Marshal(e.conf)
	finalJSON := fmt.Sprintf(`{"status":"ok","configuration":%s}`, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

// UI configuration dump API call.
func (e mainEnv) uiConfigurationDump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if len(e.conf.Notification.MagicSyncURL) != 0 &&
		len(e.conf.Notification.MagicSyncToken) != 0 {
		e.conf.UI.MagicLookup = true
	} else {
		e.conf.UI.MagicLookup = false
	}
	resultJSON, _ := json.Marshal(e.conf.UI)
	finalJSON := fmt.Sprintf(`{"status":"ok","ui":%s}`, resultJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(finalJSON))
}

// backupDB API call.
func (e mainEnv) backupDB(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if e.enforceAuth(w, r, nil) == "" {
		return
	}
	w.WriteHeader(200)
	e.db.store.BackupDB(w)
}

// setupRouter() setup HTTP Router object.
func (e mainEnv) setupRouter() *httprouter.Router {
	box := packr.NewBox("../ui")

	router := httprouter.New()

	router.GET("/v1/sys/backup", e.backupDB)
	router.GET("/v1/sys/configuration", e.configurationDump)
	router.GET("/v1/sys/uiconfiguration", e.uiConfigurationDump)

	router.POST("/v1/user", e.userNew)
	router.GET("/v1/user/:mode/:address", e.userGet)
	router.DELETE("/v1/user/:mode/:address", e.userDelete)
	router.PUT("/v1/user/:mode/:address", e.userChange)

	router.GET("/v1/login/:mode/:address", e.userLogin)
	router.GET("/v1/enter/:mode/:address/:tmp", e.userLoginEnter)

	router.POST("/v1/sharedrecord/token/:token", e.newSharedRecord)
	router.GET("/v1/get/:record", e.getRecord)

	router.GET("/v1/request/:request", e.getUserRequest)
	router.POST("/v1/request/:request", e.approveUserRequest)
	router.DELETE("/v1/request/:request", e.cancelUserRequest)
	router.GET("/v1/requests", e.getUserRequests)

	router.GET("/v1/consent/:mode/:address", e.consentAllUserRecords)
	router.GET("/v1/consent/:mode/:address/:brief", e.consentUserRecord)
	router.GET("/v1/consents/:brief", e.consentFilterRecords)
	router.GET("/v1/consents", e.consentBriefs)
	router.POST("/v1/consent/:mode/:address/:brief", e.consentAccept)
	router.DELETE("/v1/consent/:mode/:address/:brief", e.consentWithdraw)

	router.POST("/v1/userapp/token/:token/:appname", e.userappNew)
	router.GET("/v1/userapp/token/:token/:appname", e.userappGet)
	router.PUT("/v1/userapp/token/:token/:appname", e.userappChange)
	router.GET("/v1/userapp/token/:token", e.userappList)
	router.GET("/v1/userapps", e.appList)

	router.POST("/v1/session/:mode/:address", e.newSession)
	router.GET("/v1/session/:mode/:address", e.getUserSessions)

	router.GET("/v1/metrics", e.metrics)

	router.GET("/v1/audit/list/:token", e.getAuditEvents)
	router.GET("/v1/audit/get/:atoken", e.getAuditEvent)

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		data, err := box.Find("index.html")
		if err != nil {
			//log.Panic("error %s", err.Error())
			log.Printf("error: %s\n", err.Error())
			w.WriteHeader(404)
		} else {
			w.WriteHeader(200)
			w.Write([]byte(data))
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
			//w.Header().Set("Access-Control-Allow-Origin", "*")
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			}
			w.WriteHeader(200)
			w.Write([]byte(data))
		}
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("url not found"))
	})
	return router
}

// readFile() read configuration file.
func readFile(cfg *Config, filepath *string) error {
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
	e.db.expireConsentRecords(notifyURL)
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
	Code int
}

// NewCustomResponseWriter function returns CustomResponseWriter object
func NewCustomResponseWriter(ww http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		w:    ww,
		Code: 0,
	}
}

// Header function returns HTTP Header object
func (w *CustomResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

// WriteHeader function writes header back to original ResponseWriter
func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
	w.w.WriteHeader(statusCode)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		autocontext.Set(r, "host", r.Host)
		w2 := NewCustomResponseWriter(w)
		handler.ServeHTTP(w2, r)
		autocontext.Clean(r)
		log.Printf("%d %s %s\n", w2.Code, r.Method, r.URL)
	})
}

func setupDB(dbPtr *string) (*dbcon, string, error) {
	fmt.Printf("\nDatabunker init\n\n")
	masterKey, err := generateMasterKey()
	hash := md5.Sum(masterKey)
	fmt.Printf("Master key: %x\n\n", masterKey)
	fmt.Printf("Init database\n\n")
	store, err := storage.InitDB(dbPtr)
	if err != nil {
		//log.Panic("error %s", err.Error())
		log.Fatalf("db init error %s", err.Error())
	}
	db := &dbcon{store, masterKey, hash[:]}
	rootToken, err := db.createRootXtoken()
	if err != nil {
		//log.Panic("error %s", err.Error())
		fmt.Printf("error %s", err.Error())
	}
	fmt.Printf("\nAPI Root token: %s\n\n", rootToken)
	return db, rootToken, err
}

func getMasterKey(masterKeyPtr *string) []byte {
	masterKeyStr := ""
	if masterKeyPtr != nil && len(*masterKeyPtr) > 0 {
		masterKeyStr = *masterKeyPtr
	} else {
		masterKeyStr = os.Getenv("DATABUNKER_MASTERKEY")
	}
	if len(masterKeyStr) != 48 {
		fmt.Printf("Failed to decode master key: bad length\n")
		os.Exit(0)
	}
	masterKey, err := hex.DecodeString(masterKeyStr)
	if err != nil {
		fmt.Printf("Failed to decode master key: %s\n", err)
		os.Exit(0)
	}
	return masterKey
}

// main application function
func main() {
	rand.Seed(time.Now().UnixNano())
	lockMemory()
	//fmt.Printf("%+v\n", cfg)
	initPtr := flag.Bool("init", false, "generate master key and init database")
	startPtr := flag.Bool("start", false, "start databunker service. User DATABUNKER_MASTERKEY environment variable.")
	masterKeyPtr := flag.String("masterkey", "", "master key")
	dbPtr := flag.String("db", "databunker", "database file")
	confPtr := flag.String("conf", "", "configuration file name")
	flag.Parse()

	var cfg Config
	readFile(&cfg, confPtr)
	readEnv(&cfg)
	if *initPtr {
		db, _, _ := setupDB(dbPtr)
		db.store.CloseDB()
		os.Exit(0)
	}
	if storage.DBExists(dbPtr) == false {
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
	masterKey := getMasterKey(masterKeyPtr)
	store, _ := storage.OpenDB(dbPtr)
	store.InitUserApps()
	hash := md5.Sum(masterKey)
	db := &dbcon{store, masterKey, hash[:]}
	e := mainEnv{db, cfg, make(chan struct{})}
	e.dbCleanup()
	fmt.Printf("host %s\n", cfg.Server.Host+":"+cfg.Server.Port)
	router := e.setupRouter()
	srv := &http.Server{Addr: cfg.Server.Host + ":" + cfg.Server.Port, Handler: logRequest(router)}

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// Waiting for SIGINT (pkill -2)
	go func() {
		<-stop
		fmt.Println("Closing app...")
		close(e.stopChan)
		time.Sleep(1)
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
