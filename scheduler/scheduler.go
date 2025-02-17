package scheduler

import (
	"context"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

func ShouldCheckWebServices(db storage.Storage, config *conf.Config, now time.Time) bool {
	// TODO - should follow some config or smth
	return true
}

func CheckWebServices(db storage.Storage, services []conf.ServiceConfig) error {
	// TODO: using "legacy" approach to get this done quickly
	// should look into monitor.CheckWebsites refactor
	// and getting rid of WebConfig struct
	websites := make(map[string]monitor.WebConfig)
	for _, service := range services {
		// skip disabled
		if !service.Enabled {
			continue
		}
		// skip non-website services
		if service.Url == "" {
			continue
		}

		websites[service.Id] = monitor.WebConfig{
			Url:         service.Url,
			HttpStatus:  service.HttpStatus,
			BodyContent: service.BodyContent,
		}
	}
	return monitor.CheckWebsites(db, websites)
}

// Add placeholder (sentinel) "report" task to bootstrap calculation of next report time.
func InitializeSentinel(db storage.Storage, now time.Time) error {
	logger := logging.Get()
	task, err := db.LatestTaskLog("report")
	if err != nil {
		return err
	}
	// skip creation if any value already exists
	if task != nil {
		return nil
	}
	logger.Infow("Creating sentinel report task", "time", now)
	err = db.CreateTaskLog(storage.TaskInput{
		TaskName: "report", Status: string(handlers.TASK_SENTINEL), Timestamp: now, Details: ""})
	return err
}

// Decide if report should be generated and send at given time `query`.
// If error is not nil returned bool value must be ignored.
func ShouldReport(db storage.Storage, config *conf.Config, query time.Time) (bool, error) {
	task, err := db.LatestTaskLog("report")
	if err != nil {
		return false, err
	}
	// Report now if not previous report found.
	// Use InitializeSentinel to delay reporting too soon
	// after startup.
	if task == nil {
		return true, nil
	}
	// retry immediately if previous attempt failed
	if task.Status == string(handlers.TASK_ERROR) {
		return true, nil
	}

	nextReportTime := handlers.NextReportTime(config, task.Timestamp)
	isAfter := query.After(nextReportTime)
	return isAfter, nil
}

func RunSingle(db storage.Storage, config *conf.Config, now time.Time) error {
	logger := logging.Get()
	logger.Info("Do scheduling work")

	var err error = nil
	if ShouldCheckWebServices(db, config, now) {
		logger.Info("Checking web services...")
		err = CheckWebServices(db, config.AllServices())
		// TODO: might want to continue on error here
		if err != nil {
			return err
		}
	}
	doReport, err := ShouldReport(db, config, now)
	if err != nil {
		return err
	}
	if doReport {
		logger.Info("Reporting...")
		err = handlers.DoReportTask(db, config, now)
	}
	return err
}

// Run periodic jobs.
// Config: SCHEDULER_PERIOD (duration)
//
// Will not call run next job again until previous one returns, even
// if specified interval passes.
func Start(ctx context.Context, db storage.Storage, config *conf.Config) {
	logger := logging.Get()
	checkInterval := config.SchedulerPeriod
	InitializeSentinel(db, time.Now())
	logger.Infow("Starting scheduler", "checkInterval", checkInterval)
	startFunction(ctx, checkInterval, func(now time.Time) error {
		return RunSingle(db, config, now)
	})
}

func startFunction(ctx context.Context, interval time.Duration, job func(time.Time) error) {
	logger := logging.Get()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Scheduler stopped.")
			return
		case t := <-ticker.C:
			// If both ctx.Done() and ticker.C are ready
			// there is no guarantee what will be selected.
			// Hence we need to check the context was cancelled here.
			if ctx.Err() != nil {
				return
			}
			err := job(t)
			if err != nil {
				logger.Errorw("scheduled job failed", zap.Error(err))
			}
		}
	}
}
