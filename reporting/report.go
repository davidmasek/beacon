package reporting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

func GenerateReport(db storage.Storage, config *conf.Config) ([]ServiceReport, error) {
	logger := logging.Get()
	reports := make([]ServiceReport, 0)

	services := config.AllServices()

	for _, service := range services {

		healthCheck, err := db.LatestHealthCheck(service.Id)
		checks := []*storage.HealthCheck{}
		if healthCheck != nil {
			checks = append(checks, healthCheck)
		}
		var serviceStatus monitor.ServiceStatus
		if err == nil {
			serviceStatus = monitor.GetServiceStatus(service, checks)
		} else {
			logger.Errorw("error checking service status", "service", service, zap.Error(err))
			serviceStatus = monitor.STATUS_OTHER
		}
		logger.Debug("Checked service", "service", service, "status", serviceStatus)

		reports = append(reports, ServiceReport{
			ServiceStatus:     serviceStatus,
			LatestHealthCheck: healthCheck,
			ServiceCfg:        service,
		})
	}

	return reports, nil
}

// Save and send report.
//
// See ShouldReport to check if this task should be run.
func SaveSendReport(reports []ServiceReport, db storage.Storage, config *conf.Config, now time.Time) error {
	filename := config.ReportName
	if filename == "" {
		filename = "report"
	}

	if !strings.HasSuffix(filename, ".html") {
		filename = fmt.Sprintf("%s.html", filename)
	}

	// proceed to send email even if writing to file fails
	// as it is better if at least one of the two succeeds
	err := WriteReportToFile(reports, filename)

	shouldSendEmail := config.EmailConf.IsEnabled()
	if shouldSendEmail {
		emailErr := sendReport(reports, &config.EmailConf)
		err = errors.Join(err, emailErr)
	}

	status := storage.TASK_OK
	details := ""
	if err != nil {
		status = storage.TASK_ERROR
		details = err.Error()

	}
	dbErr := db.CreateTaskLog(storage.TaskInput{
		TaskName: "report", Status: string(status), Timestamp: now, Details: details})
	return errors.Join(err, dbErr)
}

func ReportFailedService(db storage.Storage, config *conf.Config, serviceCfg *conf.ServiceConfig, now time.Time) error {
	var err error
	shouldSendEmail := config.EmailConf.IsEnabled()
	prefix := config.EmailConf.Prefix
	// add whitespace after prefix if it exists and is not included already
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix = prefix + " "
	}

	msg := fmt.Sprintf(`%sBeacon: Service "%s" failed!`, prefix, serviceCfg.Id)

	if shouldSendEmail {
		err = SendMail(&config.EmailConf, msg, msg)
	}

	status := storage.TASK_OK
	if err != nil {
		status = storage.TASK_ERROR

	}
	dbErr := db.CreateTaskLog(storage.TaskInput{
		TaskName: "report_fail", Status: string(status), Timestamp: now, Details: serviceCfg.Id})
	return errors.Join(err, dbErr)
}

func RunSingle(db storage.Storage, config *conf.Config, now time.Time) error {
	logger := logging.Get()
	logger.Info("Do scheduling work")

	var err error = nil
	if scheduler.ShouldCheckWebServices(db, config, now) {
		logger.Info("Checking web services...")
		err = monitor.CheckWebServices(db, config.AllServices())
		if err != nil {
			return err
		}
	}
	doReport, err := scheduler.ShouldReport(db, config, now)
	if err != nil {
		return err
	}
	reports, err := GenerateReport(db, config)
	if err != nil {
		return err
	}
	if doReport {
		logger.Info("Reporting...")
		err = SaveSendReport(reports, db, config, now)
		if err != nil {
			return err
		}
	}
	for _, report := range reports {
		if report.ServiceStatus == monitor.STATUS_OK {
			continue
		}
		logger.Debugw("Service not OK", "service", report.ServiceCfg.Id)
		doReport, err = scheduler.ShouldReportFailedService(db, &report.ServiceCfg, now)
		if err != nil {
			return err
		}
		if doReport {
			logger.Infow("Reporting failed service", "service", report.ServiceCfg.Id)
			err = ReportFailedService(db, config, &report.ServiceCfg, now)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// Run periodic jobs.
// Config: SCHEDULER_PERIOD (duration)
//
// Run first pass immediately.
//
// Will not call run next job again until previous one returns, even
// if specified interval passes.
func Start(ctx context.Context, db storage.Storage, config *conf.Config) {
	logger := logging.Get()
	checkInterval := config.SchedulerPeriod
	err := scheduler.InitializeSentinel(db, time.Now())
	if err != nil {
		logger.Errorw("Failed to initialize job sentinel", zap.Error(err))
	}

	if err = RunSingle(db, config, time.Now()); err != nil {
		logger.Errorw("Scheduling work failed", zap.Error(err))
	}

	logger.Infow("Starting scheduler", "checkInterval", checkInterval)
	scheduler.StartFunction(ctx, checkInterval, func(now time.Time) error {
		return RunSingle(db, config, now)
	})
}
