package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/securitybunker/databunker/src/autocontext"
)

func notifyBadLogin(notifyURL string, mode string, identity string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","identity":"%s","mode":"%s"}`,
		"badlogin", identity, mode)
	host := autocontext.GetAuto("host")
	go notify(notifyURL, host, []byte(requestBody))
}

func notifyProfileNew(notifyURL string, profile []byte, mode string, identity string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","identity":"%s","mode":"%s","profile":%s}`,
		"profilenew", identity, mode, profile)
	host := autocontext.GetAuto("host")
	go notify(notifyURL, host, []byte(requestBody))
}

func notifyProfileChange(notifyURL string, old []byte, profile []byte, mode string, identity string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","identity":"%s","mode":"%s","old":%s,"profile":%s}`,
		"profilechange", identity, mode, old, profile)
	host := autocontext.GetAuto("host")
	go notify(notifyURL, host, []byte(requestBody))
}

func notifyForgetMe(notifyURL string, profile []byte, mode string, identity string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","identity":"%s","mode":"%s","profile":%s}`,
		"forgetme", identity, mode, profile)
	host := autocontext.GetAuto("host")
	go notify(notifyURL, host, []byte(requestBody))
}

func notifyConsentChange(notifyURL string, brief string, status string, mode string, identity string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody, _ := json.Marshal(map[string]string{
		"action":  "consentchange",
		"brief":   brief,
		"status":  status,
		"mode":    mode,
		"identity": identity,
	})
	host := autocontext.GetAuto("host")
	go notify(notifyURL, host, requestBody)
}

func notify(notifyURL string, host interface{}, requestBody []byte) {
	req, _ := http.NewRequest("POST", notifyURL, bytes.NewBuffer(requestBody))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if host != nil {
		req.Header.Add("Original-Host", host.(string))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error in notify: %s", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error in body read: %s", err)
		return
	}
	log.Printf("Notification result: %s", string(body))
}
