package handlers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/davidmasek/beacon/conf"
	"github.com/wneessen/go-mail"
)

func SendReport(reports []ServiceReport, emailConfig *conf.EmailConfig) error {
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
	err = SendMail(
		emailConfig,
		subject,
		buffer.String(),
	)
	return err
}

func SendMail(emailConfig *conf.EmailConfig, subject string, body string) error {
	log.Printf("Sending email with subject %q to %q", subject, emailConfig.SendTo)

	message := mail.NewMsg()
	if err := message.From(emailConfig.Sender); err != nil {
		return err
	}
	if err := message.To(emailConfig.SendTo); err != nil {
		return err
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)

	client, err := mail.NewClient(emailConfig.SmtpServer, mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(emailConfig.SmtpUsername), mail.WithPassword(emailConfig.SmtpPassword.Get()),
		mail.WithPort(emailConfig.SmtpPort),
	)
	if err != nil {
		return err
	}

	tlsSkipVerify := false
	if emailConfig.TlsInsecure == "always" {
		tlsSkipVerify = true
	} else if emailConfig.TlsInsecure == "beacon" && emailConfig.SmtpUsername == "beacon" {
		tlsSkipVerify = true
	}
	if tlsSkipVerify {
		client.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	err = client.DialAndSend(message)

	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}
	return err
}
