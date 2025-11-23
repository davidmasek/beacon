package jobs

import (
	"context"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/reporting"
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

func RunAllJobs(db storage.Storage, config *conf.Config, now time.Time) error {
	logger := logging.Get()
	logger.Info("Do scheduling work")

	err := WebCheckJob(db, config, now)
	if err != nil {
		return err
	}
	reports, err := reporting.GenerateReport(db, config)
	if err != nil {
		return err
	}
	err = reporting.SummaryReportJob(reports, db, config, now)
	if err != nil {
		return err
	}
	err = reporting.FailsReportJob(reports, db, config, now)
	if err != nil {
		return err
	}
	err = PruneDbJob(db, config, now)
	if err != nil {
		return err
	}
	return nil
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

	if err = RunAllJobs(db, config, time.Now()); err != nil {
		logger.Errorw("Scheduling work failed", zap.Error(err))
	}

	logger.Infow("Starting scheduler", "checkInterval", checkInterval)
	scheduler.StartFunction(ctx, checkInterval, func(now time.Time) error {
		return AddAll(db, config)
	})
}
