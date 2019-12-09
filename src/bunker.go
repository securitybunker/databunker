package main

import (
	"encoding/hex"
	"flag"
	"fmt"
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
	yaml "gopkg.in/yaml.v2"
)

type Tbl = int

type listTbls struct {
	Users    Tbl
	Audit    Tbl
	Xtokens  Tbl
	Consent  Tbl
	Sessions Tbl
}

// Enum for public use
var TblName = &listTbls{
	Users:    0,
	Audit:    1,
	Xtokens:  2,
	Consent:  3,
	Sessions: 4,
}

type Config struct {
	Generic struct {
		Create_user_without_token bool `yaml:"create_user_without_token"`
	}
	Sms struct {
		Default_country string `yaml:"default_country"`
		Twilio_account  string `yaml:"twilio_account"`
		Twilio_token    string `yaml:"twilio_token"`
		Twilio_from     string `yaml:"twilio_from"`
	}
	Server struct {
		Port string `yaml:"port", envconfig:"BUNKER_PORT"`
		Host string `yaml:"host", envconfig:"BUNKER_HOST"`
	} `yaml:"server"`
	Smtp struct {
		Server string `yaml:"server", envconfig:"SMTP_SERVER"`
		Port   string `yaml:"port", envconfig:"SMTP_PORT"`
		User   string `yaml:"user", envconfig:"SMTP_USER"`
		Pass   string `yaml:"pass", envconfig:"SMTP_PASS"`
		Sender string `yaml:"sender", envconfig:"SMTP_SENDER"`
	} `yaml:"smtp"`
}

type mainEnv struct {
	db   dbcon
	conf Config
}

type userJSON struct {
	jsonData []byte
	loginIdx string
	emailIdx string
	phoneIdx string
}

type tokenAuthResult struct {
	ttype   string
	name    string
	token   string
	fields  string
	appName string
}

func prometheusHandler() http.Handler {
	handlerOptions := promhttp.HandlerOpts{
		ErrorHandling:      promhttp.ContinueOnError,
		DisableCompression: true,
	}
	promHandler := promhttp.HandlerFor(prometheus.DefaultGatherer, handlerOptions)
	return promHandler
}

func (e mainEnv) metrics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Printf("/metrics\n")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	//fmt.Fprintf(w, `{"status":"ok","apps":%q}`, result)
	//fmt.Fprintf(w, "hello")
	//promhttp.Handler().ServeHTTP(w, r)
	prometheusHandler().ServeHTTP(w, r)
}

func (e mainEnv) index(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Index access\n")
	/*
		if r.Method != "GET" {
			http.Error(w, http.StatusText(405), 405)
			log.Panic("Method %s", r.Method)
			return
		}
	*/
	fmt.Fprintf(w, "<html><head><title>title</title></head></html>")
}

func (e mainEnv) setupRouter() *httprouter.Router {

	box := packr.NewBox("../ui")

	router := httprouter.New()
	router.POST("/v1/user", e.userNew)
	router.GET("/v1/user/:index/:code", e.userGet)
	router.DELETE("/v1/user/:index/:code", e.userDelete)
	router.PUT("/v1/user/:index/:code", e.userChange)

	router.GET("/v1/login/:index/:code", e.userLogin)
	router.GET("/v1/enter/:index/:code/:tmp", e.userLoginEnter)

	router.POST("/v1/xtoken/:token", e.userNewToken)
	router.GET("/v1/xtoken/:xtoken", e.userCheckToken)

	router.GET("/v1/consent/:index/:code", e.consentList)
	router.POST("/v1/consent/:index/:code", e.consentAccept)
	//router.PATCH("/v1/consent/:index/:code", e.consentCancel)
	router.DELETE("/v1/consent/:index/:code", e.consentCancel)

	router.POST("/v1/userapp/token/:token/:appname", e.userappNew)
	router.GET("/v1/userapp/token/:token/:appname", e.userappGet)
	router.PUT("/v1/userapp/token/:token/:appname", e.userappChange)
	router.GET("/v1/userapp/token/:token", e.userappList)
	router.GET("/v1/userapp/list", e.appList)

	router.GET("/v1/metrics", e.metrics)

	router.GET("/v1/audit/list/:token", e.getAuditEvents)

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		data, err := box.Find("index.html")
		if err != nil {
			//log.Panic("error %s", err.Error())
			fmt.Printf("404 %s, error: %s\n", r.URL.Path, err.Error())
			w.WriteHeader(404)
		} else {
			//fmt.Printf("return static file: %s\n", data)
			fmt.Printf("200 %s\n", r.URL.Path)
			w.WriteHeader(200)
			w.Write([]byte(data))
		}
	})
	router.GET("/site/*filepath", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		data, err := box.Find(r.URL.Path)
		if err != nil {
			fmt.Printf("404 GET %s\n", r.URL.Path)
			w.WriteHeader(404)
		} else {
			//w.Header().Set("Access-Control-Allow-Origin", "*")
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			}
			// text/plain
			fmt.Printf("200 %s\n", r.URL.Path)
			w.WriteHeader(200)
			w.Write([]byte(data))
		}
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("url not found"))
		fmt.Printf("404 %s %s\n", r.Method, r.URL.Path)
	})
	return router
}

func readFile(cfg *Config) error {
	f, err := os.Open("databunker.yaml")
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
func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	return err
}

func main() {
	rand.Seed(time.Now().UnixNano())
	lockMemory()
	var cfg Config
	readFile(&cfg)
	readEnv(&cfg)
	//fmt.Printf("%+v\n", cfg)
	initPtr := flag.Bool("init", false, "a bool")
	masterKeyPtr := flag.String("masterkey", "", "master key")
	flag.Parse()
	var err error
	var masterKey []byte
	if err != nil {
		//log.Panic("error %s", err.Error())
		fmt.Printf("error %s", err.Error())
	}
	db, _ := newDB(masterKey, nil)
	if *initPtr {
		fmt.Println("\nDatabunker init\n")
		masterKey, err = generateMasterKey()
		fmt.Printf("Master key: %x\n\n", masterKey)
		fmt.Println("Init databunker.db\n")
		db.initDB()
		rootToken, err := db.createRootToken()
		if err != nil {
			//log.Panic("error %s", err.Error())
			fmt.Printf("error %s", err.Error())
		}
		fmt.Printf("API Root token: %s\n\n", rootToken)
		os.Exit(0)
	}
	if dbExists() == false {
		fmt.Println("\ndatabunker.db file is missing.\n")
		fmt.Println(`Run "./databunker -init" for the first time.`)
		fmt.Println("")
		os.Exit(0)
	}
	if masterKeyPtr != nil && len(*masterKeyPtr) > 0 {
		masterKey, err = hex.DecodeString(*masterKeyPtr)
	} else {
		fmt.Println(`Masterkey is missing.`)
		fmt.Println(`Run "./databunker -masterkey key"`)
		os.Exit(0)
	}
	db.initUserApps()
	e := mainEnv{db, cfg}
	fmt.Printf("host %s\n", cfg.Server.Host+":"+cfg.Server.Port)
	router := e.setupRouter()
	if _, err := os.Stat("./server.key"); !os.IsNotExist(err) {
		//TODO
		fmt.Printf("Loading ssl\n")
		err := http.ListenAndServeTLS(":443", "server.ctr", "server.key", router)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		log.Fatal(http.ListenAndServe(cfg.Server.Host+":"+cfg.Server.Port, router))
	}
}
