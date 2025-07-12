package jobs

import (
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
)

func WebCheckJob(db storage.Storage, config *conf.Config, now time.Time) error {
	logger := logging.Get()
	doCheckWeb, err := scheduler.ShouldCheckWebServices(db, config, now)
	if err != nil {
		return err
	}
	if doCheckWeb {
		logger.Info("Checking web services...")
		err = monitor.CheckWebServices(db, config.AllServices())
		if err != nil {
			return err
		}
	}
	return nil
}
