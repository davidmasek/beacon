package jobs

import (
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
)

// ~two years
const Retention = 2 * 365 * 24 * time.Hour

func PruneDbJob(db storage.Storage, cfg *conf.Config, now time.Time) error {
	logger := logging.Get()
	cutoff := now.Add(-Retention)
	rows, err := db.PruneHealthChecks(cutoff)
	if err != nil {
		return err
	}
	logger.Infow("pruned health checks", "rows", rows, "cutoff", cutoff)

	rows, err = db.PruneTasks(cutoff)
	if err != nil {
		return err
	}
	logger.Infow("pruned tasks", "rows", rows, "cutoff", cutoff)
	return nil
}
