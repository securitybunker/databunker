package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"
)

func sendCodeByEmail(code int32, identity string, cfg Config) {
	Dest := []string{identity}
	Subject := "Access Code"
	bodyMessage := "Access code is " + strconv.Itoa(int((code)))
	msg := "From: " + cfg.SMTP.Sender + "\n" +
		"To: " + strings.Join(Dest, ",") + "\n" +
		"Subject: " + Subject + "\n" + bodyMessage
	auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Server)
	err := smtp.SendMail(cfg.SMTP.Server+":"+cfg.SMTP.Port,
		auth, cfg.SMTP.User, Dest, []byte(msg))
	if err != nil {
		log.Printf("error sending email: %s", err)
		return
	}
	log.Printf("Mail sent successfully!")
}

func adminEmailAlert(action string, adminEmail string, cfg Config) {
	if len(adminEmail) == 0 {
		return
	}
	Dest := []string{adminEmail}
	Subject := "Data Subject request received"
	bodyMessage := "Request: " + action
	msg := "From: " + cfg.SMTP.Sender + "\n" +
		"To: " + strings.Join(Dest, ",") + "\n" +
		"Subject: " + Subject + "\n" + bodyMessage
	auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Server)
	err := smtp.SendMail(cfg.SMTP.Server+":"+cfg.SMTP.Port,
		auth, cfg.SMTP.User, Dest, []byte(msg))
	if err != nil {
		fmt.Printf("smtp error: %s", err)
		return
	}
	fmt.Println("Mail sent successfully!")
}
