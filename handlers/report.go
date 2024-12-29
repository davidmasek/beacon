package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

func GenerateReport(db storage.Storage) ([]ServiceReport, error) {
	reports := make([]ServiceReport, 0)

	services, err := db.ListServices()
	if err != nil {
		return nil, err
	}

	checkConfig := ServiceChecker{
		Timeout: 24 * time.Hour,
	}

	for _, service := range services {
		log.Println("Checking service", service)
		healthCheck, err := db.LatestHealthCheck(service)
		if err != nil {
			// TODO: should probably still include in the report with some explanation
			log.Println("[ERROR]", err)
			continue
		}
		serviceStatus, err := checkConfig.GetServiceStatus(healthCheck)
		if err != nil {
			// TODO: should probably still include in the report with some explanation
			log.Println("[ERROR]", err)
			continue
		}
		log.Println(" - Service status:", serviceStatus)

		reports = append(reports, ServiceReport{
			ServiceId: service, ServiceStatus: serviceStatus, LatestHealthCheck: healthCheck,
		})
	}

	return reports, nil
}

func sendEmail(config *viper.Viper, reports []ServiceReport) error {
	if !config.IsSet("email") {
		err := fmt.Errorf("no email configuration provided")
		return err
	}
	server, err := LoadServer(config.Sub("email"))
	if err != nil {
		err := fmt.Errorf("failed to load SMTP server: %w", err)
		return err
	}
	mailer := SMTPMailer{
		Server: server,
	}
	return mailer.Send(reports, config.Sub("email"))
}

// Generate, save and send report.
//
// See ShouldReport to check if this task should be run.
func DoReportTask(db storage.Storage, config *viper.Viper, now time.Time) error {
	reports, err := GenerateReport(db)
	if err != nil {
		return err
	}

	config.SetDefault("report-name", "report")
	filename := config.GetString("report-name")

	if !strings.HasSuffix(filename, ".html") {
		filename = fmt.Sprintf("%s.html", filename)
	}

	// proceed to send email even if writing to file fails
	// as it is better if at least one of the two succeeds
	err = WriteReportToFile(reports, filename)

	// TODO: need better way so check if should send email
	// currently the config.IsSet handles the "config-file path"
	// and allows overwrite via config "send-mail" variable for CLI usage
	shouldSendEmail := config.IsSet("email.smtp_password")
	if config.IsSet("send-mail") {
		shouldSendEmail = config.GetBool("send-mail")
	}
	if shouldSendEmail {
		emailErr := sendEmail(config, reports)
		err = errors.Join(err, emailErr)
	}

	status := "OK"
	if err != nil {
		status = err.Error()
	}
	dbErr := db.CreateTaskLog("report", status, now)
	return errors.Join(err, dbErr)
}
