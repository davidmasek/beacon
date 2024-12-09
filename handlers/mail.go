package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/spf13/viper"
)

type FakeMailer struct {
	Target string
}

func (fm FakeMailer) Name() string {
	return "FakeMailer"
}

func (fm FakeMailer) Handle(reports []ServiceReport) error {
	// Create or truncate the output file
	filename := "report.html"
	log.Printf("[%s] Writing report to %s", fm.Name(), filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = WriteReport(reports, file)
	if err != nil {
		return err
	}

	log.Printf("[%s] Saved to %s. Would be send to %q", fm.Name(), filename, fm.Target)
	return nil
}

type SMTPMailer struct {
	Server SMTPServer
	Target string
	Env    string
}

func (sm SMTPMailer) Name() string {
	return "SMTPMailer"
}

func (sm SMTPMailer) Handle(reports []ServiceReport) error {
	var buffer bytes.Buffer

	log.Printf("[%s] Generating report", sm.Name())
	err := WriteReport(reports, &buffer)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("[%s] Heartbeat report", sm.Env)
	log.Printf("[%s] Sending email with subject %q to %q", sm.Name(), subject, sm.Target)
	err = SendMail(sm.Server, subject, buffer.String())
	if err != nil {
		log.Printf("[%s] Failed to send email: %v", sm.Name(), err)
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
func LoadServer(viper *viper.Viper) (SMTPServer, error) {
	// TODO: error handling
	return SMTPServer{
		server:   viper.GetString("SMTP_SERVER"),
		port:     viper.GetString("SMTP_PORT"),
		username: viper.GetString("SMTP_USERNAME"),
		password: viper.GetString("SMTP_PASSWORD"),
	}, nil
}

func SendMail(server SMTPServer, subject string, body string) error {
	// Email details
	sender := "noreply@optimisticotter.me"
	recipient := "davidmasek42@gmail.com"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\""

	// Format the email message
	msg := "From: " + sender + "\n" +
		"To: " + recipient + "\n" +
		"Subject: " + subject + "\n" +
		mime + "\n\n" + body

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", server.username, server.password, server.server)
	err := smtp.SendMail(server.server+":"+server.port, auth, sender, []string{recipient}, []byte(msg))
	return err
}
