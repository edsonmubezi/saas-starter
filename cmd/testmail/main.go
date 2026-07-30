package main

import (
	"fmt"
	"gopkg.in/gomail.v2"
)

func main() {
	d := gomail.NewDialer("smtp.gmail.com", 587, "edsondominic80@gmail.com", "slvgkkdvdbrhjufu")

	m := gomail.NewMessage()
	m.SetHeader("From", "Microfinance <edsondominic80@gmail.com>")
	m.SetHeader("To", "edsondominic80@gmail.com")
	m.SetHeader("Subject", "Microfinance - SMTP Test")
	m.SetBody("text/plain", "Hello!\n\nThis is a test email from Microfinance to verify SMTP is working.\n\nIf you received this, email sending is configured correctly.")

	fmt.Println("Sending test email to edsondominic80@gmail.com ...")
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("FAILED: %v\n", err)
	} else {
		fmt.Println("SUCCESS: Email sent!")
	}
}
