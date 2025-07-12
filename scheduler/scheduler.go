package scheduler

import (
	"context"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

// report daily (might be made configurable later)
const FailedServiceReportInterval = 24 * time.Hour

func ShouldCheckWebServices(db storage.Storage, config *conf.Config, now time.Time) (bool, error) {
	task, err := db.LatestTaskLog("web_check")
	if err != nil {
		return false, err
	}
	if task == nil {
		return true, nil
	}
	// retry immediately if previous attempt failed
	if task.Status == string(storage.TASK_ERROR) {
		return true, nil
	}
	nextReportTime := task.Timestamp.Add(config.WebCheckPeriod)
	isAfter := now.After(nextReportTime)
	return isAfter, nil
}

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
func NextReportTime(config *conf.Config, lastReportTime time.Time) time.Time {
	lastReportTime = lastReportTime.In(config.Timezone.Location)
	nextReportDay := lastReportTime.AddDate(0, 0, 1)
	// if ReportOnDays is set, then keep adding days until we get one that is allowed
	for !config.ReportOnDays.IsEmpty() && !config.ReportOnDays.Contains(nextReportDay) {
		nextReportDay = nextReportDay.AddDate(0, 0, 1)
	}
	nextReportTime := time.Date(
		nextReportDay.Year(), nextReportDay.Month(), nextReportDay.Day(),
		config.ReportAfter, 0, 0, 0, nextReportDay.Location())
	return nextReportTime
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
	if task.Status == string(storage.TASK_ERROR) {
		return true, nil
	}

	nextReportTime := NextReportTime(config, task.Timestamp)
	isAfter := query.After(nextReportTime)
	return isAfter, nil
}

// Decide if extra report should be generated for a failed service
func ShouldReportFailedService(db storage.Storage, cfg *conf.ServiceConfig, query time.Time) (bool, error) {
	task, err := db.LatestServiceFailedLog(cfg.Id)
	if err != nil {
		return false, err
	}
	if task == nil {
		return true, nil
	}
	if task.Status == string(storage.TASK_ERROR) {
		return true, nil
	}
	nextReportTime := task.Timestamp.Add(FailedServiceReportInterval)
	isAfter := query.After(nextReportTime)
	return isAfter, nil
}

func StartFunction(ctx context.Context, interval time.Duration, job func(time.Time) error) {
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
		TaskName: "report", Status: string(storage.TASK_SENTINEL), Timestamp: now, Details: ""})
	return err
}
