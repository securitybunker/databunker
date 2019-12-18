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
	Dest := []string{"stremovsky@gmail.com", address}
	Subject := "Access Code"
	bodyMessage := "Data bunker access code is " + strconv.Itoa(int((code)))
	msg := "From: " + cfg.Smtp.Sender + "\n" +
		"To: " + strings.Join(Dest, ",") + "\n" +
		"Subject: " + Subject + "\n" + bodyMessage

	err := smtp.SendMail(cfg.Smtp.Server+":"+cfg.Smtp.Port,
		smtp.PlainAuth("", cfg.Smtp.User, cfg.Smtp.Pass, cfg.Smtp.Server),
		cfg.Smtp.User, Dest, []byte(msg))

	if err != nil {
		fmt.Printf("smtp error: %s", err)
		return
	}

	fmt.Println("Mail sent successfully!")
}
