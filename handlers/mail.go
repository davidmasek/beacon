package handlers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

func SendReport(reports []ServiceReport, emailConfig *conf.EmailConfig) error {
	var buffer bytes.Buffer
	logger := logging.Get()
	logger.Info("Generating report")
	err := WriteReport(reports, &buffer)
	if err != nil {
		return err
	}

	prefix := emailConfig.Prefix
	// add whitespace after prefix if it exists and is not included already
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}

	nServices := len(reports)
	nGood := 0
	for _, report := range reports {
		if report.ServiceStatus == monitor.STATUS_OK {
			nGood += 1
		}
	}

	statusSummary := "Status Report"
	if nServices == nGood {
		statusSummary = "All Good"
	} else {
		statusSummary = "Service(s) Failed"
	}

	subject := fmt.Sprintf("%sBeacon: %s [%d/%d]", prefix, statusSummary, nGood, nServices)
	err = SendMail(
		emailConfig,
		subject,
		buffer.String(),
	)
	return err
}

func SendMail(emailConfig *conf.EmailConfig, subject string, body string) error {
	logger := logging.Get()
	logger.Infow("Sending email", "subject", subject, "to", emailConfig.SendTo)

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

	if emailConfig.SmtpSSL {
		client.SetSSL(true)
	}

	err = client.DialAndSend(message)

	if err != nil {
		logger.Errorw("Failed to send email", "subject", subject, "to", emailConfig.SendTo, zap.Error(err))
	}
	return err
}
