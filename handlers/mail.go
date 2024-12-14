package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Mailer interface {
	Send(reports []ServiceReport, config *viper.Viper) error
}

// TODO/feature
// TODO: re-fit FakeMailer to smth like "FileReport"
// use as default and use SMTP as extra if there is the --send-mail flag
type FakeMailer struct{}

func (fm FakeMailer) Send(reports []ServiceReport, config *viper.Viper) error {
	// Create or truncate the output file
	filename := "report.html"
	log.Printf("[FakeMailer] Writing report to %s", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = WriteReport(reports, file)
	if err != nil {
		return err
	}

	target := config.GetString("send_to")

	log.Printf("[FakeMailer] Saved to %s. Would be send to %q", filename, target)
	return nil
}

type SMTPMailer struct {
	Server SMTPServer
}

func (sm SMTPMailer) Send(reports []ServiceReport, config *viper.Viper) error {
	var buffer bytes.Buffer

	log.Printf("[SMTPMailer] Generating report")
	err := WriteReport(reports, &buffer)
	if err != nil {
		return err
	}

	config.SetDefault("prefix", "")
	prefix := config.GetString("prefix")
	// add whitespace after prefix if it exists and is not included already
	if prefix != "" && strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}

	subject := fmt.Sprintf("%sBeacon: Status Report", prefix)
	target := config.GetString("send_to")
	log.Printf("[SMTPMailer] Sending email with subject %q to %q", subject, target)
	err = SendMail(
		sm.Server,
		config.GetString("sender"),
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

// Load the SMTP server details from the .env file
func LoadServer(config *viper.Viper) (SMTPServer, error) {
	// TODO: error handling
	return SMTPServer{
		server:   config.GetString("SMTP_SERVER"),
		port:     config.GetString("SMTP_PORT"),
		username: config.GetString("SMTP_USERNAME"),
		password: config.GetString("SMTP_PASSWORD"),
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
