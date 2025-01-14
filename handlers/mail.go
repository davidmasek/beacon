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

func (sm SMTPMailer) Send(reports []ServiceReport, emailConfig *conf.EmailConfig) error {
	var buffer bytes.Buffer

	log.Printf("[SMTPMailer] Generating report")
	err := WriteReport(reports, &buffer)
	if err != nil {
		return err
	}

	prefix := emailConfig.Prefix
	// add whitespace after prefix if it exists and is not included already
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}

	subject := fmt.Sprintf("%sBeacon: Status Report", prefix)
	log.Printf("[SMTPMailer] Sending email with subject %q to %q", subject, emailConfig.SendTo)
	err = SendMail(
		sm.Server,
		emailConfig.Sender,
		emailConfig.SendTo,
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
	password conf.Secret
}

// Load the SMTP server details from config
func LoadServer(emailConfig *conf.EmailConfig) SMTPServer {
	return SMTPServer{
		server:   emailConfig.SmtpServer,
		port:     fmt.Sprint(emailConfig.SmtpPort),
		username: emailConfig.SmtpUsername,
		password: emailConfig.SmtpPassword,
	}
}

func SendMail(server SMTPServer, sender string, target string, subject string, body string) error {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\""

	// Format the email message
	msg := "From: " + sender + "\n" +
		"To: " + target + "\n" +
		"Subject: " + subject + "\n" +
		mime + "\n\n" + body

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", server.username, server.password.Get(), server.server)
	err := smtp.SendMail(server.server+":"+server.port, auth, sender, []string{target}, []byte(msg))
	return err
}
