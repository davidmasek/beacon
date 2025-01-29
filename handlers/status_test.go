package handlers

import (
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/require"
)

func TestGetServiceStatus(t *testing.T) {
	checkConfig := ServiceChecker{
		Timeout: 24 * time.Hour,
	}

	hc := &storage.HealthCheck{
		Id:        -1,
		ServiceId: "some-service",
		Timestamp: time.Now(),
		Metadata:  map[string]string{},
	}

	status := checkConfig.GetServiceStatus(hc)
	require.Equal(t, monitor.STATUS_OK, status)

	hc.Timestamp = hc.Timestamp.Add(-time.Hour)
	status = checkConfig.GetServiceStatus(hc)
	require.Equal(t, monitor.STATUS_OK, status)

	hc.Timestamp = hc.Timestamp.Add(-30 * time.Hour)
	status = checkConfig.GetServiceStatus(hc)
	require.Equal(t, monitor.STATUS_FAIL, status)
}

func TestGetServiceStatusWithError(t *testing.T) {
	checkConfig := ServiceChecker{
		Timeout: 24 * time.Hour,
	}

	status := checkConfig.GetServiceStatus(&storage.HealthCheck{
		Id:        -1,
		ServiceId: "some-service",
		Timestamp: time.Time{},
		Metadata: map[string]string{
			"error": "much bad",
		},
	})

	require.Equal(t, monitor.STATUS_FAIL, status)
}
