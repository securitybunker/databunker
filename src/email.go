package main

import (
	"log"
	"net/smtp"
	"strconv"
	"strings"
)

func sendCodeByEmail(code int32, identity string, cfg Config) {
	dest := []string{identity}
	target := strings.Join(dest, ",")
	subject := "Access Code"
	bodyMessage := "Access code is " + strconv.Itoa(int((code)))
	msg := "From: " + cfg.SMTP.Sender + "\n" +
		"To: " + target + "\n" +
		"Subject: " + subject + "\n" +
		bodyMessage
	auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Server)
	err := smtp.SendMail(cfg.SMTP.Server+":"+cfg.SMTP.Port,
		auth, cfg.SMTP.User, dest, []byte(msg))
	log.Printf("Send email to %s via %s", target, cfg.SMTP.Server)
	if err != nil {
		log.Printf("Error sending email: %s", err)
		return
	}
	log.Printf("Mail sent successfully!")
}

func adminEmailAlert(action string, adminEmail string, cfg Config) {
	if len(adminEmail) == 0 {
		return
	}
	dest := []string{adminEmail}
	subject := "Data Subject request received"
	bodyMessage := "Request: " + action
	msg := "From: " + cfg.SMTP.Sender + "\n" +
		"To: " + strings.Join(dest, ",") + "\n" +
		"Subject: " + subject + "\n" +
		bodyMessage
	auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Server)
	err := smtp.SendMail(cfg.SMTP.Server+":"+cfg.SMTP.Port,
		auth, cfg.SMTP.User, dest, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
	log.Println("Mail sent successfully!")
}
