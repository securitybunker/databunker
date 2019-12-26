package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func notifyConsent(notifyUrl string, brief string, status string, mode string, address string) {
	formData := url.Values{
		"action":  {"consentchange"},
		"brief":   {brief},
		"status":  {status},
		"mode":    {mode},
		"address": {address},
	}
	resp, err := http.PostForm(notifyUrl, formData)
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
	log.Printf("Consent notification result: %s", string(body))
}
