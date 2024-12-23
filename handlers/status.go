package handlers

import (
	"log"
	"time"

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
func (config *ServiceChecker) GetServiceStatus(latestHealthCheck *storage.HealthCheck) (monitor.ServiceStatus, error) {
	if latestHealthCheck == nil {
		log.Println("[GetServiceStatus] no health check found")
		return monitor.STATUS_FAIL, nil
	}
	timeAgo := time.Since(latestHealthCheck.Timestamp)
	log.Printf("timeout: %s, timeAgo: %s", config.Timeout.String(), timeAgo.String())
	if timeAgo > config.Timeout {
		return monitor.STATUS_FAIL, nil
	}

	if errorMeta, exists := latestHealthCheck.Metadata["error"]; exists {
		if errorMeta != "" {
			log.Printf("[GetServiceStatus] error found: %q", errorMeta)
			return monitor.STATUS_FAIL, nil
		}
	}
	if statusMeta, exists := latestHealthCheck.Metadata["status"]; exists {
		if statusMeta != string(monitor.STATUS_OK) {
			log.Printf("[GetServiceStatus] status not OK: %q != %q", statusMeta, string(monitor.STATUS_OK))
			return monitor.STATUS_FAIL, nil
		}
	}
	return monitor.STATUS_OK, nil
}
