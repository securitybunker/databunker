package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ttacon/libphonenumber"
	"golang.org/x/sys/unix"
)

var (
	regexUUID       = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	regexAppName    = regexp.MustCompile("^[a-z][a-z0-9]{1,20}$")
	regexExpiration = regexp.MustCompile("^([0-9]+)([mhds])$")
)

func normalizeEmail(email0 string) string {
	email, _ := url.QueryUnescape(email0)
	email = strings.ToLower(email)
	email = strings.TrimSpace(email)
	if email0 != email {
		fmt.Printf("email before: %s, after: %s\n", email0, email)
	}
	return email
}

func normalizePhone(phone string, default_country string) string {
	// 4444 is a phone number for testing, no need to normilize it
	phone = strings.TrimSpace(phone)
	if phone == "4444" {
		return "4444"
	}
	if len(default_country) == 0 {
		default_country = "UK"
	}
	res, err := libphonenumber.Parse(phone, default_country)
	if err != nil {
		fmt.Printf("failed to parse phone number: %s", phone)
		return ""
	}
	phone = "+" + strconv.Itoa(int(*res.CountryCode)) + strconv.FormatUint(*res.NationalNumber, 10)
	return phone
}

func validateMode(index string) bool {
	if index == "token" {
		return true
	}
	if index == "email" {
		return true
	}
	if index == "phone" {
		return true
	}
	if index == "login" {
		return true
	}
	return false
}

func parseFields(fields string) []string {
	return strings.Split(fields, ",")
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func atoi(s string) int32 {
	var (
		n uint32
		i int
		v byte
	)
	for ; i < len(s); i++ {
		d := s[i]
		if '0' <= d && d <= '9' {
			v = d - '0'
		} else if 'a' <= d && d <= 'z' {
			v = d - 'a' + 10
		} else if 'A' <= d && d <= 'Z' {
			v = d - 'A' + 10
		} else {
			n = 0
			break
		}
		n *= uint32(10)
		n += uint32(v)
	}
	return int32(n)
}

func parseExpiration(expiration string) (int32, error) {
	match := regexExpiration.FindStringSubmatch(expiration)
	// expiration format: 10d, 10h, 10m, 10s
	if len(match) != 3 {
		e := fmt.Sprintf("failed to parse expiration value: %s", expiration)
		return 0, errors.New(e)
	}
	num := match[1]
	format := match[2]
	start := int32(time.Now().Unix())
	switch format {
	case "d":
		start = start + (atoi(num) * 24 * 3600)
	case "h":
		start = start + (atoi(num) * 3600)
	case "m":
		start = start + (atoi(num) * 60)
	case "s":
		start = start + (atoi(num))
	}
	return start, nil
}

func lockMemory() error {
	return unix.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)
}

func isValidUUID(uuidCode string) bool {
	return regexUUID.MatchString(uuidCode)
}

func isValidApp(app string) bool {
	return regexAppName.MatchString(app)
}

func returnError(w http.ResponseWriter, r *http.Request, message string, code int, err error, event *auditEvent) {
	fmt.Printf("%d %s %s\n", code, r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"status":%q,"message":%q}`, "error", message)
	if event != nil {
		event.Status = "error"
		event.Msg = message
		if err != nil {
			event.Debug = err.Error()
			fmt.Printf("Msg: %s, Error: %s\n", message, err.Error())
		} else {
			fmt.Printf("Msg: %s\n", message)
		}
	}
	//http.Error(w, http.StatusText(405), 405)
}

func returnUUID(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":%q}`, code)
}

func (e mainEnv) enforceAuth(w http.ResponseWriter, r *http.Request, event *auditEvent) bool {
	/*
		for key, value := range r.Header {
			fmt.Printf("%s => %s\n", key, value)
		}
	*/
	if token, ok := r.Header["X-Bunker-Token"]; ok {
		authResult, err := e.db.checkUserAuthXToken(token[0])
		//fmt.Printf("error in auth? error %s - %s\n", err, token[0])
		if err == nil {
			if event != nil {
				event.Who = authResult.name
			}
			if authResult.ttype == "login" {
				if authResult.token == event.Record {
					return true
					// else go down in code
				}
			} else {
				return true
			}
		}
		/*
			if e.db.checkToken(token[0]) == true {
				if event != nil {
					event.Who = "admin"
				}
				return true
			}
		*/
	}
	fmt.Printf("403 Access denied\n")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Access denied"))
	if event != nil {
		event.Status = "error"
		event.Msg = "access denied"
	}
	return false
}

func enforceUUID(w http.ResponseWriter, uuidCode string, event *auditEvent) bool {
	if isValidUUID(uuidCode) == false {
		fmt.Printf("405 bad uuid in : %s\n", uuidCode)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintf(w, `{"status":"error","message":"bad uuid"}`)
		if event != nil {
			event.Status = "error"
			event.Msg = "bad uuid"
		}
		return false
	}
	return true
}

func getJSONPostData(r *http.Request) (map[string]interface{}, error) {
	cType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	cType = strings.ToLower(cType)
	records := make(map[string]interface{})
	//body, _ := ioutil.ReadAll(r.Body)
	//fmt.Printf("Body: %s\n", body)
	if r.Method == "DELETE" {
		// other wise data is not parsed!
		r.Method = "PATCH"
	}

	if strings.HasPrefix(cType, "application/x-www-form-urlencoded") {
		err = r.ParseForm()
		if err != nil {
			fmt.Printf("error in http data parsing: %s\n", err)
			return nil, err
		}
		for key, value := range r.Form {
			//fmt.Printf("data here %s => %s\n", key, value[0])
			records[key] = value[0]
		}
	} else if strings.HasPrefix(cType, "application/json") {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &records)
		if err != nil {
			return nil, err
		}
	} else {
		e := fmt.Sprintf("wrong content type: %s", cType)
		return nil, errors.New(e)
	}
	return records, nil
}

func getJSONPost(r *http.Request, default_country string) (userJSON, error) {
	records, err := getJSONPostData(r)
	var result userJSON

	if value, ok := records["login"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			result.loginIdx = value.(string)
			result.loginIdx = strings.TrimSpace(result.loginIdx)
		}
	}
	if value, ok := records["email"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			result.emailIdx = value.(string)
			result.emailIdx = normalizeEmail(result.emailIdx)
		}
	}
	if value, ok := records["phone"]; ok {
		if reflect.TypeOf(value) == reflect.TypeOf("string") {
			result.phoneIdx = value.(string)
			result.phoneIdx = normalizePhone(result.phoneIdx, default_country)
		}
	}

	result.jsonData, err = json.Marshal(records)

	return result, err
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var numbers = []rune("0123456789")

func randNum(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}
