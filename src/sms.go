package databunker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func sendCodeByPhone(code int32, address string, cfg Config) {
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + cfg.Sms.TwilioAccount + "/Messages.json"
	fmt.Printf("url %s\n", urlStr)
	msgData := url.Values{}
	msgData.Set("To", address)
	msgData.Set("From", cfg.Sms.TwilioFrom)
	msgData.Set("Body", "Data Bunker code "+strconv.Itoa(int(code)))
	msgDataReader := *strings.NewReader(msgData.Encode())
	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(cfg.Sms.TwilioAccount, cfg.Sms.TwilioToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
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
