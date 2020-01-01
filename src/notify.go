package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func notifyProfileNew(notifyUrl string, profile []byte, mode string, address string) {
	if len(notifyUrl) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","profile":%s}`,
		"profilenew", address, mode, profile)
	go notify(notifyUrl, []byte(requestBody))
}

func notifyProfileChange(notifyUrl string, old []byte, profile []byte, mode string, address string) {
	if len(notifyUrl) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","old":%s,"profile":%s}`,
		"profilechange", address, mode, old, profile)
	go notify(notifyUrl, []byte(requestBody))
}

func notifyForgetMe(notifyUrl string, profile []byte, mode string, address string) {
	if len(notifyUrl) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","profile":%s}`,
		"forgetme", address, mode, profile)
	go notify(notifyUrl, []byte(requestBody))
}

func notifyConsentChange(notifyUrl string, brief string, status string, mode string, address string) {
	if len(notifyUrl) == 0 {
		return
	}
	requestBody, _ := json.Marshal(map[string]string{
		"action":  "consentchange",
		"brief":   brief,
		"status":  status,
		"mode":    mode,
		"address": address,
	})
	go notify(notifyUrl, requestBody)
}

func notify(notifyUrl string, requestBody []byte) {
	resp, err := http.Post(notifyUrl, "application/json", bytes.NewBuffer(requestBody))
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
