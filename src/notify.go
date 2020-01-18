package databunker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func notifyProfileNew(notifyURL string, profile []byte, mode string, address string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","profile":%s}`,
		"profilenew", address, mode, profile)
	go notify(notifyURL, []byte(requestBody))
}

func notifyProfileChange(notifyURL string, old []byte, profile []byte, mode string, address string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","old":%s,"profile":%s}`,
		"profilechange", address, mode, old, profile)
	go notify(notifyURL, []byte(requestBody))
}

func notifyForgetMe(notifyURL string, profile []byte, mode string, address string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody := fmt.Sprintf(`{"action":"%s","address":"%s","mode":"%s","profile":%s}`,
		"forgetme", address, mode, profile)
	go notify(notifyURL, []byte(requestBody))
}

func notifyConsentChange(notifyURL string, brief string, status string, mode string, address string) {
	if len(notifyURL) == 0 {
		return
	}
	requestBody, _ := json.Marshal(map[string]string{
		"action":  "consentchange",
		"brief":   brief,
		"status":  status,
		"mode":    mode,
		"address": address,
	})
	go notify(notifyURL, requestBody)
}

func notify(notifyURL string, requestBody []byte) {
	resp, err := http.Post(notifyURL, "application/json", bytes.NewBuffer(requestBody))
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
