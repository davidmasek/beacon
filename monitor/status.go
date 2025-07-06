package monitor

import (
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
)

type ServiceStatus string

const (
	STATUS_OK   ServiceStatus = "OK"
	STATUS_FAIL ServiceStatus = "FAIL"
	// e.g. unable to decide, not enough data, error in the check
	STATUS_OTHER ServiceStatus = "OTHER"
)

func HealthCheckStatus(hc *storage.HealthCheck) ServiceStatus {
	logger := logging.Get()
	if errorMeta, exists := hc.Metadata["error"]; exists {
		if errorMeta != "" {
			logger.Debugf("error found: %q", errorMeta)
			return STATUS_FAIL
		}
	}
	if statusMeta, exists := hc.Metadata["status"]; exists {
		if statusMeta != string(STATUS_OK) {
			logger.Debugf("status not OK: %q != %q", statusMeta, string(STATUS_OK))
			return STATUS_FAIL
		}
	}
	return STATUS_OK
}

func GetServiceStatus(serviceCfg conf.ServiceConfig, checks []*storage.HealthCheck) ServiceStatus {
	logger := logging.Get()
	logger.Debugw("get service status", "service", serviceCfg, "checks", checks)
	if len(checks) == 0 {
		logger.Debug("no health check found")
		return STATUS_FAIL
	}
	latestHealthCheck := checks[len(checks)-1]
	timeAgo := time.Since(latestHealthCheck.Timestamp)
	logger.Debugw("service status", "timeout", serviceCfg.Timeout.String(), "timeAgo", timeAgo.String())
	if timeAgo > serviceCfg.Timeout {
		return STATUS_FAIL
	}
	return HealthCheckStatus(latestHealthCheck)
}
