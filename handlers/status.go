package handlers

import (
	"time"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

type ServiceChecker struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

func DefaultServiceChecker() ServiceChecker {
	return ServiceChecker{
		Timeout: 24 * time.Hour,
	}
}

// Compute service status based on latest HealthCheck
//
// TODO: maybe should be used like HealthCheck.GetStatus(config) or smth
func (config *ServiceChecker) GetServiceStatus(latestHealthCheck *storage.HealthCheck) monitor.ServiceStatus {
	logger := logging.Get()
	if latestHealthCheck == nil {
		logger.Debug("[GetServiceStatus] no health check found")
		return monitor.STATUS_FAIL
	}
	timeAgo := time.Since(latestHealthCheck.Timestamp)
	logger.Debug("timeout: %s, timeAgo: %s", config.Timeout.String(), timeAgo.String())
	if timeAgo > config.Timeout {
		return monitor.STATUS_FAIL
	}

	if errorMeta, exists := latestHealthCheck.Metadata["error"]; exists {
		if errorMeta != "" {
			logger.Debug("[GetServiceStatus] error found: %q", errorMeta)
			return monitor.STATUS_FAIL
		}
	}
	if statusMeta, exists := latestHealthCheck.Metadata["status"]; exists {
		if statusMeta != string(monitor.STATUS_OK) {
			logger.Debug("[GetServiceStatus] status not OK: %q != %q", statusMeta, string(monitor.STATUS_OK))
			return monitor.STATUS_FAIL
		}
	}
	return monitor.STATUS_OK
}
