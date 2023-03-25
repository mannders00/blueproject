package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

const (
	smtpHost     = "smtp.office365.com"
	smtpPort     = 587
	smtpUsername = "support@blueproject.info"
	smtpPassword = "tdet8Z!wzW"
)

func sendVerificationEmail(dest, token string) {
	subject := "BlueProject Email Verification"
	verificationURL := fmt.Sprintf("http://localhost:8080/verify/%s", token)
	body := fmt.Sprintf("Please click the following link to verify your email: %s", verificationURL)

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", dest)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	err := d.DialAndSend(m)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Email sent")
	}
}

func generateToken() string {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes)
}
