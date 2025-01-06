package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/davidmasek/beacon/conf"
)

type SMTPMailer struct {
	Server SMTPServer
}

func (sm SMTPMailer) Send(reports []ServiceReport, emailConfig *conf.Config) error {
	var buffer bytes.Buffer

	log.Printf("[SMTPMailer] Generating report")
	err := WriteReport(reports, &buffer)
	if err != nil {
		return err
	}

	prefix := ""
	if emailConfig.IsSet("prefix") {
		prefix = emailConfig.GetString("prefix")
	}
	// add whitespace after prefix if it exists and is not included already
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}

	subject := fmt.Sprintf("%sBeacon: Status Report", prefix)
	target := emailConfig.GetString("send_to")
	log.Printf("[SMTPMailer] Sending email with subject %q to %q", subject, target)
	err = SendMail(
		sm.Server,
		emailConfig.GetString("sender"),
		target,
		subject,
		buffer.String(),
	)
	if err != nil {
		log.Printf("[SMTPMailer] Failed to send email: %v", err)
		return err
	}

	return nil
}

type SMTPServer struct {
	server   string
	port     string
	username string
	password string
}

// Load the SMTP server details from config
func LoadServer(emailConfig *conf.Config) (SMTPServer, error) {
	return SMTPServer{
		server:   emailConfig.GetString("SMTP_SERVER"),
		port:     emailConfig.GetString("SMTP_PORT"),
		username: emailConfig.GetString("SMTP_USERNAME"),
		password: emailConfig.GetString("SMTP_PASSWORD"),
	}, nil
}

func SendMail(server SMTPServer, sender string, target string, subject string, body string) error {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\""

	// Format the email message
	msg := "From: " + sender + "\n" +
		"To: " + target + "\n" +
		"Subject: " + subject + "\n" +
		mime + "\n\n" + body

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", server.username, server.password, server.server)
	err := smtp.SendMail(server.server+":"+server.port, auth, sender, []string{target}, []byte(msg))
	return err
}
