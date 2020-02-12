package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func sendCodeByPhone(code int32, address string, cfg Config) {
	domain := "https://api.twilio.com"
	client := &http.Client{}
	sendCodeByPhoneDo(domain, client, code, address, cfg)
}

func sendCodeByPhoneDo(domain string, client *http.Client, code int32, address string, cfg Config) {
	urlStr := domain + "/2010-04-01/Accounts/" + cfg.Sms.TwilioAccount + "/Messages.json"
	fmt.Printf("url %s\n", urlStr)
	msgData := url.Values{}
	msgData.Set("To", address)
	msgData.Set("From", cfg.Sms.TwilioFrom)
	msgData.Set("Body", "Data Bunker code "+strconv.Itoa(int(code)))
	msgDataReader := *strings.NewReader(msgData.Encode())
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(cfg.Sms.TwilioAccount, cfg.Sms.TwilioToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error sending sms")
		return
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}
}
