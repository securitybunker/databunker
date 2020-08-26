package main

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

func sendCodeByEmail(code int32, address string, cfg Config) {
	/*
		c, err := smtp.Dial(smtpServer)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		// Set the sender and recipient.
		c.Mail("bot@paranoidguy.com")
		c.Rcpt(address)
		// Send the email body.
		wc, err := c.Data()
		if err != nil {
			log.Fatal(err)
			return
		}
		defer wc.Close()
		buf := bytes.NewBufferString("This is the email body.")
		if _, err = buf.WriteTo(wc); err != nil {
			log.Fatal(err)
			return
		}
		return
	*/
	Dest := []string{address}
	Subject := "Access Code"
	bodyMessage := "Access code is " + strconv.Itoa(int((code)))
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
