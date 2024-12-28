// Package main - Personal Identifiable Information (PII) database.
// For more info check https://databunker.org/
package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/securitybunker/databunker/src/audit"
	"github.com/securitybunker/databunker/src/autocontext"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v2"
)

var version string

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
		DisableAudit                 bool   `yaml:"disable_audit" default:"false"`
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
	if e.EnforceAuth(w, r, nil) == "" {
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

// ReadConfFile() read configuration file.
func ReadConfFile(cfg *Config, filepath *string) error {
	confFile := "databunker.yaml"
	if filepath != nil {
		if len(*filepath) > 0 {
			confFile = *filepath
		}
	}
	log.Printf("Loading configuration file: %s\n", confFile)
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
func ReadEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	return err
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

	router.POST("/v1/sharedrecord/:mode/:identity", e.newSharedRecord)
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

	router.POST("/v1/userapp/:mode/:identity/:appname", e.userappNew)
	router.GET("/v1/userapp/:mode/:identity/:appname", e.userappGet)
	router.PUT("/v1/userapp/:mode/:identity/:appname", e.userappChange)
	router.DELETE("/v1/userapp/:mode/:identity/:appname", e.userappDelete)
	router.GET("/v1/userapp/:mode/:identity", e.userappList)
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
	router.GET("/robots.txt", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		data, err := box.Find("robots.txt")
		if err != nil {
			//log.Panic("error %s", err.Error())
			log.Printf("error: %s\n", err.Error())
			w.WriteHeader(404)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write(data)
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

// dbCleanup() is used to run cron jobs.
func (e mainEnv) dbCleanupDo() {
	log.Printf("db cleanup timeout\n")
	exp, _ := utils.ParseExpiration0(e.conf.Policy.MaxAuditRetentionPeriod)
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

// helper function to load user details by idex name
func (e mainEnv) loadUserToken(w http.ResponseWriter, r *http.Request, mode string, identity string, event *audit.AuditEvent) string {
	var err error
	if utils.ValidateMode(mode) == false {
		utils.ReturnError(w, r, "bad mode", 405, nil, event)
		return ""
	}
	var userBson bson.M
	if mode == "token" {
		if utils.EnforceUUID(w, identity, event) == false {
			return ""
		}
		userBson, err = e.db.lookupUserRecord(identity)
	} else {
		userBson, err = e.db.lookupUserRecordByIndex(mode, identity, e.conf)
	}
	if userBson == nil || err != nil {
		utils.ReturnError(w, r, "internal error", 405, nil, event)
		return ""
	}
	event.Record = userBson["token"].(string)
	return event.Record
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

func (e mainEnv) reqMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("Set host %s\n", r.Host)
		e.initContext(r)
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

// main application function
func main() {
	rand.Seed(time.Now().UnixNano())
	utils.LockMemory()
	loadService()
}
