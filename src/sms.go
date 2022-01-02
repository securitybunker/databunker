package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func sendCodeByPhone(code int32, identity string, cfg Config) {
	domain := "https://api.twilio.com"
	client := &http.Client{}
	sendCodeByPhoneDo(domain, client, code, identity, cfg)
}

func sendCodeByPhoneDo(domain string, client *http.Client, code int32, identity string, cfg Config) {
	if len(cfg.Sms.URL) == 0 {
		log.Printf("SMS gateway provider URL is missing")
		return
	}
	msg := "Databunker code " + strconv.Itoa(int(code))
	finalURL := cfg.Sms.URL
	finalURL = strings.ReplaceAll(finalURL, "_PHONE_", url.QueryEscape(identity))
	finalURL = strings.ReplaceAll(finalURL, "_FROM_", url.QueryEscape(cfg.Sms.From))
	finalURL = strings.ReplaceAll(finalURL, "_TOKEN_", url.QueryEscape(cfg.Sms.Token))
	finalURL = strings.ReplaceAll(finalURL, "_MSG_", url.QueryEscape(msg))
	fmt.Printf("finalURL: %s\n", finalURL)
	if len(cfg.Sms.Method) == 0 || strings.ToUpper(cfg.Sms.Method) == "GET" {
		req, _ := http.NewRequest("GET", finalURL, nil)
		if len(cfg.Sms.BasicAuth) > 0 && strings.Contains(cfg.Sms.BasicAuth, ":") {
			s := strings.SplitN(cfg.Sms.BasicAuth, ":", 2)
			if len(s) == 2 {
				req.SetBasicAuth(strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
			}
		}
		if len(cfg.Sms.CustomHeader) > 0 && strings.Contains(cfg.Sms.CustomHeader, ":") {
			s := strings.SplitN(cfg.Sms.CustomHeader, ":", 2)
			if len(s) == 2 {
				req.Header.Add(strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
			}
		}
		resp, _ := client.Do(req)
		fmt.Println(resp.Status)
		return
	}
	body := cfg.Sms.Body
	if len(body) == 0 {
		log.Printf("Body can not be empty when performing POST request.")
		return
	}
	cType := cfg.Sms.ContentType
	if cType == "json" || cType == "application/json" {
		// no need to escape values when sending JSON
		body = strings.ReplaceAll(body, "_FROM_", cfg.Sms.From)
		body = strings.ReplaceAll(body, "_PHONE_", identity)
		body = strings.ReplaceAll(body, "_TOKEN_", cfg.Sms.Token)
		body = strings.ReplaceAll(body, "_MSG_", msg)
		cType = "application/json"
	} else {
		body = strings.ReplaceAll(body, "_FROM_", url.QueryEscape(cfg.Sms.From))
		body = strings.ReplaceAll(body, "_PHONE_", url.QueryEscape(identity))
		body = strings.ReplaceAll(body, "_TOKEN_", url.QueryEscape(cfg.Sms.Token))
		body = strings.ReplaceAll(body, "_MSG_", url.QueryEscape(msg))
		cType = "application/x-www-form-urlencoded"
	}
	//urlStr := domain + "/2010-04-01/Accounts/" + cfg.Sms.TwilioAccount + "/Messages.json"
	msgDataReader := *strings.NewReader(body)
	req, _ := http.NewRequest("POST", finalURL, &msgDataReader)
	if len(cfg.Sms.BasicAuth) > 0 && strings.Contains(cfg.Sms.BasicAuth, ":") {
		s := strings.SplitN(cfg.Sms.BasicAuth, ":", 2)
		if len(s) == 2 {
			req.SetBasicAuth(strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
		}
	}
	if len(cfg.Sms.CustomHeader) > 0 && strings.Contains(cfg.Sms.CustomHeader, ":") {
		s := strings.SplitN(cfg.Sms.CustomHeader, ":", 2)
		if len(s) == 2 {
			req.Header.Add(strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
		}
	}
	req.Header.Add("Content-Type", cType)
	resp, _ := client.Do(req)
	fmt.Println(resp.Status)
}
