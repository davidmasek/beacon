package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/storage"
)

// Calculate when the next report should happen based on last report time.
// No previous reports or failed reporting tasks are not considered here.
// See `ShouldReport` for more complex logic.
//
// Since timezones are hard, there might be some edge cases with weird behavior.
// Suppose lastReportTime is 2025-01-01 01:00 in Brisbane (UTC+10)
// and we want the next report time in Honolulu (UTC-10).
// The result will be 2025-01-01 09:00:00 -1000 HST (in Pacific/Honolulu).
// Is that what you expect? Note that adding the 24 hours first and
// then converting to (target) local time leads to different result.
//
// TODO: refactor packages
// - it does not make sense for this one report scheduling function to be here
// if all the others are in scheduler
// - moving it to scheduler creates circular dependency
// - probably makes more sense to move stuff from scheduler here anyway
// - but at that moment `handlers` is too big of a package (which it is anyway by now)
// -> split handlers, create:
// - reporting/report
// - server / web_gui / smth
// - keep scheduler with reduced functionality ?
func NextReportTime(config *conf.Config, lastReportTime time.Time) time.Time {
	lastReportTime = lastReportTime.In(&config.Timezone)
	nextReportDay := lastReportTime.Add(24 * time.Hour)
	nextReportTime := time.Date(
		nextReportDay.Year(), nextReportDay.Month(), nextReportDay.Day(),
		config.ReportAfter, 0, 0, 0, nextReportDay.Location())
	return nextReportTime
}

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

func sendEmail(config *conf.Config, reports []ServiceReport) error {
	if !config.IsSet("email") {
		err := fmt.Errorf("no email configuration provided")
		return err
	}
	server, err := LoadServer(&config.EmailConf)
	if err != nil {
		err := fmt.Errorf("failed to load SMTP server: %w", err)
		return err
	}
	mailer := SMTPMailer{
		Server: server,
	}
	return mailer.Send(reports, &config.EmailConf)
}

// Generate, save and send report.
//
// See ShouldReport to check if this task should be run.
func DoReportTask(db storage.Storage, config *conf.Config, now time.Time) error {
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
	shouldSendEmail := config.EmailConf.SmtpPassword.IsSet()
	if config.IsSet("send-mail") {
		shouldSendEmail = config.GetBool("send-mail")
	}
	if shouldSendEmail {
		emailErr := sendEmail(config, reports)
		err = errors.Join(err, emailErr)
	}

	status := TASK_OK
	details := ""
	if err != nil {
		status = TASK_ERROR
		details = err.Error()

	}
	dbErr := db.CreateTaskLog(storage.TaskInput{
		TaskName: "report", Status: string(status), Timestamp: now, Details: details})
	return errors.Join(err, dbErr)
}
