package controller

import (
	"fmt"
	"net/smtp"
)

type Emailer struct {
	emailAddr		string
	emailPassword	string
	smtpHost		string
	smtpPort		string
}

func NewEmailer(emailAddr string, emailPassword string, smtpHost string, smtpPort string) *Emailer {
	var emailer Emailer
	emailer.emailAddr = emailAddr
	emailer.emailPassword = emailPassword
	emailer.smtpHost = smtpHost
	emailer.smtpPort = smtpPort
	return &emailer
}

func (emailer *Emailer) SendEmail(destEmail []string, subject string, message string) error {
	auth := smtp.PlainAuth("", emailer.emailAddr, emailer.emailPassword, emailer.smtpHost)

	fromStr := fmt.Sprintf("From: <%s>\r\n", "hallo@calendariumculinarium.de")
	toStr := fmt.Sprintf("To: <%s>\r\n", "sebastian@slowfoodyouthh.de")
	subjectStr := "Subject: " + subject + "\r\n"
	body := message + "\r\n"

	msg := []byte(fromStr+toStr+subjectStr+"\r\n"+body)

	err := smtp.SendMail(emailer.smtpHost + ":" + emailer.smtpPort, auth, emailer.emailAddr, destEmail, msg)
	return err
}
