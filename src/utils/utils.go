package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
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
	regexBrief      = regexp.MustCompile("^[a-z][a-z0-9\\-]{1,64}$")
	regexAppName    = regexp.MustCompile("^[a-z][a-z0-9\\_]{1,30}$")
	regexExpiration = regexp.MustCompile("^([0-9]+)([mhds])?$")
	regexHex        = regexp.MustCompile("^[a-zA-F0-9]+$")
	letters         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numbers         = []rune("0123456789")
	numbers0        = []rune("123456789")
	indexNames      = map[string]bool{
		"custom": true,
		"email":  true,
		"login":  true,
		"phone":  true,
		"token":  true,
	}
	consentValueTypes = map[string]bool{
		"1":       true,
		"accept":  true,
		"agree":   true,
		"approve": true,
		"given":   true,
		"good":    true,
		"ok":      true,
		"on":      true,
		"true":    true,
		"y":       true,
		"yes":     true,
	}
	legalBasisTypes = map[string]bool{
		"consent":             true,
		"contract":            true,
		"legal-requirement":   true,
		"legitimate-interest": true,
		"public-interest":     true,
		"vital-interest":      true,
	}
)

// UserJSON used to parse user POST
type UserJSONStruct struct {
	JsonData  []byte
	LoginIdx  string
	EmailIdx  string
	PhoneIdx  string
	CustomIdx string
	Token     string
}

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
		log.Printf("Checking header: %s\n", idx0)
		if utils.SliceContains(interestingHeaders, idx0) {
			headers[idx] = val[0]
		}
	}
	headersStr, _ := json.Marshal(headers)
	meta := fmt.Sprintf(`{"clientip":"%s","headers":%s}`, r.RemoteAddr, headersStr)
	log.Printf("Meta: %s\n", meta)
	return meta
}
*/

func GetStringValue(r interface{}) string {
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

func GetIntValue(r interface{}) int {
	switch r.(type) {
	case int:
		return r.(int)
	case int32:
		return int(r.(int32))
	case float64:
		return int(r.(float64))
	}
	return 0
}

func GetInt64Value(records map[string]interface{}, key string) int64 {
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

func GetArgEnvFileVariable(vname string, masterKeyPtr *string) string {
	strvalue := ""
	if masterKeyPtr != nil && len(*masterKeyPtr) > 0 {
		strvalue = *masterKeyPtr
	} else if len(os.Getenv(vname)) > 0 {
		strvalue = os.Getenv(vname)
	} else if len(os.Getenv(vname+"_FILE")) > 0 {
		data, _ := os.ReadFile(os.Getenv(vname + "_FILE"))
		strvalue = string(data)
	}
	return strings.TrimSpace(strvalue)
}

func HashString(md5Salt []byte, src string) string {
	stringToHash := append(md5Salt, []byte(src)...)
	hashed := sha256.Sum256(stringToHash)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

func NormalizeConsentValue(status string) string {
	status = strings.ToLower(status)
	if consentValueTypes[status] {
		return "yes"
	}
	return "no"
}

func NormalizeLegalBasisType(status string) string {
	status = strings.ToLower(status)
	if legalBasisTypes[status] {
		return status
	}
	return "consent"
}

func NormalizeBrief(brief string) string {
	return strings.ToLower(brief)
}

func NormalizeEmail(email0 string) string {
	email, _ := url.QueryUnescape(email0)
	email = strings.ToLower(email)
	email = strings.TrimSpace(email)
	if email0 != email {
		log.Printf("Email before normalization: %s, after: %s\n", email0, email)
	}
	return email
}

func NormalizePhone(phone string, defaultCountry string) string {
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
		log.Printf("Failed to parse phone number: %s", phone)
		return ""
	}
	phone = "+" + strconv.Itoa(int(*res.CountryCode)) + strconv.FormatUint(*res.NationalNumber, 10)
	return phone
}

func ValidateMode(index string) bool {
	return indexNames[strings.ToLower(index)]
}

func ParseFields(fields string) []string {
	return strings.Split(fields, ",")
}

func SliceContains(slice []string, item string) bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set[item]
}

func Atoi(s string) int32 {
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

func SetExpiration(maxExpiration string, userExpiration string) string {
	if len(userExpiration) == 0 {
		return maxExpiration
	}
	userExpirationNum, _ := ParseExpiration(userExpiration)
	maxExpirationNum, _ := ParseExpiration(maxExpiration)
	if maxExpirationNum == 0 {
		maxExpiration = "6m"
		maxExpirationNum, _ = ParseExpiration(maxExpiration)
	}
	if userExpirationNum == 0 {
		return maxExpiration
	}
	if userExpirationNum > maxExpirationNum {
		return maxExpiration
	}
	return userExpiration
}

func ParseExpiration0(expiration string) (int32, error) {
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
		start = start + (Atoi(num) * 24 * 3600)
	case "h": // hour
		start = start + (Atoi(num) * 3600)
	case "m": // month
		start = start + (Atoi(num) * 24 * 31 * 3600)
	case "s":
		start = start + (Atoi(num))
	}
	return start, nil
}

func ParseExpiration(expiration string) (int32, error) {
	match := regexExpiration.FindStringSubmatch(expiration)
	// expiration format: 10d, 10h, 10m, 10s
	if len(match) == 2 {
		return Atoi(match[1]), nil
	}
	if len(match) != 3 {
		e := fmt.Sprintf("failed to parse expiration value: %s", expiration)
		return 0, errors.New(e)
	}
	num := match[1]
	format := match[2]
	if len(format) == 0 {
		return Atoi(num), nil
	}
	start := int32(time.Now().Unix())
	switch format {
	case "d": // day
		start = start + (Atoi(num) * 24 * 3600)
	case "h": // hour
		start = start + (Atoi(num) * 3600)
	case "m": // month
		start = start + (Atoi(num) * 24 * 31 * 3600)
	case "s":
		start = start + (Atoi(num))
	}
	return start, nil
}

func LockMemory() error {
	return unix.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)
}

func CheckValidUUID(uuidCode string) bool {
	return regexUUID.MatchString(uuidCode)
}

func CheckValidApp(app string) bool {
	return regexAppName.MatchString(app)
}

func CheckValidBrief(brief string) bool {
	return regexBrief.MatchString(brief)
}

func CheckValidHex(hex1 string) bool {
	return regexHex.MatchString(hex1)
}

// StringPatternMatch looks for basic human patterns like "*", "*abc*", etc...
func StringPatternMatch(pattern string, value string) bool {
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

func ReturnUUID(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"status":"ok","token":%q}`, code)
}

func GetJSONPostMap(r *http.Request) (map[string]interface{}, error) {
	cType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ignoring empty content-type: %s\n", err)
		return nil, nil
	}
	cType = strings.ToLower(cType)
	records := make(map[string]interface{})
	if r.Method == "DELETE" {
		// otherwise data is not parsed!
		r.Method = "PATCH"
	}
	body0, err := io.ReadAll(r.Body)
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
			log.Printf("error to parse HTTP data request: %s\n", err)
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

func GetJSONPostData(r *http.Request) ([]byte, error) {
	cType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("ignoring empty content-type: %s\n", err)
		return nil, nil
	}
	cType = strings.ToLower(cType)
	if r.Method == "DELETE" {
		// otherwise data is not parsed!
		r.Method = "PATCH"
	}
	body0, err := io.ReadAll(r.Body)
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
			log.Printf("error in HTTP data request: %s\n", err)
			return nil, err
		}
		if len(form) == 0 {
			return nil, nil
		}
		records := make(map[string]interface{})
		for key, value := range form {
			records[key] = value[0]
		}
		return json.Marshal(records)
	} else if strings.HasPrefix(cType, "application/json") || strings.HasPrefix(cType, "application/xml") {
		var data interface{}
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			return nil, errors.New("error decoding json data")
		}
		return json.Marshal(data)
	}
	log.Printf("Ignore wrong content type: %s", cType)
	maxStrLen := 200
	if len(body) < maxStrLen {
		maxStrLen = len(body)
	}
	log.Printf("Body[max 200 chars]: %s", body[0:maxStrLen])
	return nil, errors.New("wrong content-type, not a json string")
}

func GetIndexString(val interface{}) string {
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

func GetUserJSONStruct(r *http.Request, defaultCountry string) (UserJSONStruct, error) {
	var result UserJSONStruct
	records, err := GetJSONPostMap(r)
	if err != nil {
		return result, err
	}
	if records == nil {
		return result, nil
	}

	if value, ok := records["login"]; ok {
		result.LoginIdx = GetIndexString(value)
	}
	if value, ok := records["email"]; ok {
		result.EmailIdx = NormalizeEmail(GetIndexString(value))
	}
	if value, ok := records["phone"]; ok {
		result.PhoneIdx = NormalizePhone(GetIndexString(value), defaultCountry)
	}
	if value, ok := records["custom"]; ok {
		result.CustomIdx = GetIndexString(value)
	}
	if value, ok := records["token"]; ok {
		result.Token = value.(string)
	}
	result.JsonData, err = json.Marshal(records)
	return result, err
}

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandNum(n int) int32 {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	b[0] = numbers0[rand.Intn(len(numbers0))]
	return Atoi(string(b))
}
