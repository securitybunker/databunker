package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ttacon/libphonenumber"
	"golang.org/x/sys/unix"
)

var (
	regexUUID          = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	regexBrief         = regexp.MustCompile("^[a-z][a-z0-9\\-]{1,64}$")
	regexAppName       = regexp.MustCompile("^[a-z][a-z0-9\\_]{1,30}$")
	regexExpiration    = regexp.MustCompile("^([0-9]+)([mhds])?$")
	regexHex           = regexp.MustCompile("^[a-zA-F0-9]+$")
	consentYesStatuses = []string{"y", "yes", "accept", "agree", "approve", "given", "true", "good"}
	basisTypes         = []string{"consent", "contract", "legitimate-interest", "vital-interest", "legal-requirement", "public-interest"}
)

// Consideration why collection of meta data patch was postpone:
// 1. Databunker is not anti-fraud solution
// 2. GDPR stands for data minimalization.
// 3. Do not store what you actually do not NEED.
/*
var interestingHeaders = []string{"x-forwarded", "x-forwarded-for", "x-coming-from", "via",
	"forwarded-for", "forwarded", "client-ip", "user-agent", "cookie", "referer"}

func getMeta(r *http.Request) string {
	headers := bson.M{}
	for idx, val := range r.Header {
		idx0 := strings.ToLower(idx)
		fmt.Printf("checking header: %s\n", idx0)
		if contains(interestingHeaders, idx0) {
			headers[idx] = val[0]
		}
	}
	headersStr, _ := json.Marshal(headers)
	meta := fmt.Sprintf(`{"clientip":"%s","headers":%s}`, r.RemoteAddr, headersStr)
	fmt.Printf("Meta: %s\n", meta)
	return meta
}
*/

func getStringValue(r interface{}) string {
	if r == nil {
		return ""
	}
	switch r.(type) {
	case string:
		return strings.TrimSpace(r.(string))
	case []uint8:
		return strings.TrimSpace(string(r.([]uint8)))
	case float64:
	        return strconv.Itoa(int(r.(float64)))
	}
	return ""
}

func getIntValue(r interface{}) int {
	switch r.(type) {
	case int:
		return r.(int)
	case int32:
		return int(r.(int32))
	}
	return 0
}

func getInt64Value(records map[string]interface{}, key string) int64 {
	if value, ok := records[key]; ok {
		switch value.(type) {
		case int32:
			return int64(value.(int32))
		case int64:
			return int64(value.(int64))
		}
	}
	return 0
}

func hashString(md5Salt []byte, src string) string {
	stringToHash := append(md5Salt, []byte(src)...)
	hashed := sha256.Sum256(stringToHash)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

func normalizeConsentStatus(status string) string {
	status = strings.ToLower(status)
	if contains(consentYesStatuses, status) {
		return "yes"
	}
	return "no"
}

func normalizeBasisType(status string) string {
	status = strings.ToLower(status)
	if contains(basisTypes, status) {
		return status
	}
	return "consent"
}

func normalizeBrief(brief string) string {
	return strings.ToLower(brief)
}

func normalizeEmail(email0 string) string {
	email, _ := url.QueryUnescape(email0)
	email = strings.ToLower(email)
	email = strings.TrimSpace(email)
	if email0 != email {
		fmt.Printf("email before: %s, after: %s\n", email0, email)
	}
	return email
}

func normalizePhone(phone string, defaultCountry string) string {
	// 4444 is a phone number for testing, no need to normilize it
	phone = strings.TrimSpace(phone)
	if len(phone) == 0 {
		return phone
	}
	if phone == "4444" {
		return "4444"
	}
	if len(defaultCountry) == 0 {
		// https://github.com/ttacon/libphonenumber/blob/master/countrycodetoregionmap.go
		defaultCountry = "GB"
	}
	res, err := libphonenumber.Parse(phone, defaultCountry)
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
	if index == "custom" {
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

func setExpiration(maxExpiration string, userExpiration string) string {
	if len(userExpiration) == 0 {
		return maxExpiration
	}
	userExpirationNum, _ := parseExpiration(userExpiration)
	maxExpirationNum, _ := parseExpiration(maxExpiration)
	if maxExpirationNum == 0 {
		maxExpiration = "6m"
		maxExpirationNum, _ = parseExpiration(maxExpiration)
	}
	if userExpirationNum == 0 {
		return maxExpiration
	}
	if userExpirationNum > maxExpirationNum {
		return maxExpiration
	}
	return userExpiration
}

func parseExpiration0(expiration string) (int32, error) {
	match := regexExpiration.FindStringSubmatch(expiration)
	// expiration format: 10d, 10h, 10m, 10s
	if len(match) != 3 {
		e := fmt.Sprintf("failed to parse expiration value: %s", expiration)
		return 0, errors.New(e)
	}
	num := match[1]
	format := match[2]
	start := int32(0)
	switch format {
	case "d": // day
		start = start + (atoi(num) * 24 * 3600)
	case "h": // hour
		start = start + (atoi(num) * 3600)
	case "m": // month
		start = start + (atoi(num) * 24 * 31 * 3600)
	case "s":
		start = start + (atoi(num))
	}
	return start, nil
}

func parseExpiration(expiration string) (int32, error) {
	match := regexExpiration.FindStringSubmatch(expiration)
	// expiration format: 10d, 10h, 10m, 10s
	if len(match) == 2 {
		return atoi(match[1]), nil
	}
	if len(match) != 3 {
		e := fmt.Sprintf("failed to parse expiration value: %s", expiration)
		return 0, errors.New(e)
	}
	num := match[1]
	format := match[2]
	if len(format) == 0 {
		return atoi(num), nil
	}
	start := int32(time.Now().Unix())
	switch format {
	case "d": // day
		start = start + (atoi(num) * 24 * 3600)
	case "h": // hour
		start = start + (atoi(num) * 3600)
	case "m": // month
		start = start + (atoi(num) * 24 * 31 * 3600)
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

func isValidBrief(brief string) bool {
	return regexBrief.MatchString(brief)
}

func isValidHex(hex1 string) bool {
	return regexHex.MatchString(hex1)
}

// stringPatternMatch looks for basic human patterns like "*", "*abc*", etc...
func stringPatternMatch(pattern string, value string) bool {
	if len(pattern) == 0 {
		return false
	}
	if pattern == "*" {
		return true
	}
	if pattern == value {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		pattern = pattern[1 : len(pattern)-1]
		if strings.Contains(value, pattern) {
			return true
		}
		return false
	}
	if strings.HasPrefix(pattern, "*") {
		pattern = pattern[1:]
		if strings.HasSuffix(value, pattern) {
			return true
		}
	} else if strings.HasSuffix(pattern, "*") {
		pattern = pattern[:len(pattern)-1]
		if strings.HasPrefix(value, pattern) {
			return true
		}
	}
	return false
}

func returnError(w http.ResponseWriter, r *http.Request, message string, code int, err error, event *auditEvent) {
	fmt.Printf("%d %s %s\n", code, r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"status":"error","message":%q}`, message)
	if event != nil {
		event.Status = "error"
		event.Msg = message
		if err != nil {
			event.Debug = err.Error()
			log.Printf("Generate error response: %s, Error: %s\n", message, err.Error())
		} else {
			log.Printf("Generate error response: %s\n", message)
		}
	}
	//http.Error(w, http.StatusText(405), 405)
}

func returnUUID(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":%q}`, code)
}

func (e mainEnv) enforceAuth(w http.ResponseWriter, r *http.Request, event *auditEvent) string {
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
				event.Identity = authResult.name
				if authResult.ttype == "login" && authResult.token == event.Record {
					return authResult.ttype
				}
			}
			if len(authResult.ttype) > 0 && authResult.ttype != "login" {
				return authResult.ttype
			}
		}
		/*
			if e.db.checkXtoken(token[0]) == true {
				if event != nil {
					event.Identity = "admin"
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
	return ""
}

func (e mainEnv) enforceAdmin(w http.ResponseWriter, r *http.Request) string {
	if token, ok := r.Header["X-Bunker-Token"]; ok {
		authResult, err := e.db.checkUserAuthXToken(token[0])
		//fmt.Printf("error in auth? error %s - %s\n", err, token[0])
		if err == nil {
			if len(authResult.ttype) > 0 && authResult.ttype != "login" {
				return authResult.ttype
			}
		}
	}
	fmt.Printf("403 Access denied\n")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Access denied"))
	return ""
}

func enforceUUID(w http.ResponseWriter, uuidCode string, event *auditEvent) bool {
	if isValidUUID(uuidCode) == false {
		//fmt.Printf("405 bad uuid in : %s\n", uuidCode)
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

func getJSONPostMap(r *http.Request) (map[string]interface{}, error) {
	cType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		fmt.Printf("ignoring empty content-type: %s\n", err)
		return nil, nil
	}
	cType = strings.ToLower(cType)
	records := make(map[string]interface{})
	if r.Method == "DELETE" {
		// otherwise data is not parsed!
		r.Method = "PATCH"
	}
	body0, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(string(body0))
	if len(body) < 3 {
		return nil, nil
	}
	if strings.HasPrefix(cType, "application/x-www-form-urlencoded") {
		if body[0] == '{' {
			return nil, errors.New("wrong content-type, json instead of url encoded data")
		}
		form, err := url.ParseQuery(body)
		if err != nil {
			fmt.Printf("error in http data parsing: %s\n", err)
			return nil, err
		}
		if len(form) == 0 {
			return nil, nil
		}
		for key, value := range form {
			//fmt.Printf("data here %s => %s\n", key, value[0])
			records[key] = value[0]
		}
	} else if strings.HasPrefix(cType, "application/json") {
		err = json.Unmarshal([]byte(body), &records)
		if err != nil {
			log.Printf("Error in json decode %s", err)
			return nil, err
		}
	} else if strings.HasPrefix(cType, "application/xml") {
		err = json.Unmarshal([]byte(body), &records)
		if err != nil {
			log.Printf("Error in xml/json decode %s", err)
			return nil, err
		}
	} else {
		log.Printf("Ignore wrong content type: %s", cType)
		maxStrLen := 200
		if len(body) < maxStrLen {
			maxStrLen = len(body)
		}
		log.Printf("Body[max 200 chars]: %s", body[0:maxStrLen])
		return nil, nil
	}
	return records, nil
}

func getJSONPostData(r *http.Request) ([]byte, error){
	cType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		fmt.Printf("ignoring empty content-type: %s\n", err)
		return nil, nil
	}
	cType = strings.ToLower(cType)
	records := make(map[string]interface{})
	var records2 []map[string]interface{}
	if r.Method == "DELETE" {
		// otherwise data is not parsed!
		r.Method = "PATCH"
	}
	body0, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(string(body0))
	if len(body) < 3 {
		return nil, nil
	}
	if strings.HasPrefix(cType, "application/x-www-form-urlencoded") {
		if body[0] == '{' || body[0] == '[' {
			return nil, errors.New("wrong content-type, json instead of url encoded data")
		}
		form, err := url.ParseQuery(body)
		if err != nil {
			fmt.Printf("error in http data parsing: %s\n", err)
			return nil, err
		}
		if len(form) == 0 {
			return nil, nil
		}
		for key, value := range form {
			records[key] = value[0]
		}
		return json.Marshal(records)
	} else if strings.HasPrefix(cType, "application/json") || strings.HasPrefix(cType, "application/xml") {
		if body[0] == '{' {
			err = json.Unmarshal([]byte(body), &records)
		} else if body[0] == '[' {
			err = json.Unmarshal([]byte(body), &records2)
		} else {
			return nil, errors.New("wrong content-type, not a json string")
		}
		if err != nil {
			return nil, err
		}
		if body[0] == '{' {
			return json.Marshal(records)
	        } else if body[0] == '[' {
			return json.Marshal(records2)
	        }
	}
	log.Printf("Ignore wrong content type: %s", cType)
	maxStrLen := 200
	if len(body) < maxStrLen {
		maxStrLen = len(body)
	}
	log.Printf("Body[max 200 chars]: %s", body[0:maxStrLen])
	return nil, errors.New("wrong content-type, not a json string")
}

func getIndexString(val interface{}) string {
	switch val.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(val.(string))
	case []uint8:
		return strings.TrimSpace(string(val.([]uint8)))
	case int:
		return strconv.Itoa(val.(int))
	case int64:
		return fmt.Sprintf("%v", val.(int64))
	case float64:
		return strconv.Itoa(int(val.(float64)))
	}
	return ""
}

func getJSONPost(r *http.Request, defaultCountry string) (userJSON, error) {
	var result userJSON
	records, err := getJSONPostMap(r)
	if err != nil {
		return result, err
	}
	if records == nil {
		return result, nil
	}

	if value, ok := records["login"]; ok {
		result.loginIdx = getIndexString(value)
	}
	if value, ok := records["email"]; ok {
		result.emailIdx = normalizeEmail(getIndexString(value))
	}
	if value, ok := records["phone"]; ok {
		result.phoneIdx = normalizePhone(getIndexString(value), defaultCountry)
	}
	if value, ok := records["custom"]; ok {
		result.customIdx = getIndexString(value)
	}
	if value, ok := records["token"]; ok {
		result.token = value.(string)
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

var numbers0 = []rune("123456789")
var numbers = []rune("0123456789")

func randNum(n int) int32 {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	b[0] = numbers0[rand.Intn(len(numbers0))]
	return atoi(string(b))
}
