package handlers

import (
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/require"
)

func TestGetServiceStatus(t *testing.T) {
	hc := &storage.HealthCheck{
		Id:        -1,
		ServiceId: "some-service",
		Timestamp: time.Now(),
		Metadata:  map[string]string{},
	}
	checks := []*storage.HealthCheck{
		hc,
	}
	cfg := conf.ServiceConfig{
		Timeout: 24 * time.Hour,
	}

	status := monitor.GetServiceStatus(cfg, checks)
	require.Equal(t, monitor.STATUS_OK, status)

	hc.Timestamp = hc.Timestamp.Add(-time.Hour)
	status = monitor.GetServiceStatus(cfg, checks)
	require.Equal(t, monitor.STATUS_OK, status)

	hc.Timestamp = hc.Timestamp.Add(-30 * time.Hour)
	status = monitor.GetServiceStatus(cfg, checks)
	require.Equal(t, monitor.STATUS_FAIL, status)
}

func TestGetServiceStatusWithError(t *testing.T) {
	hc := &storage.HealthCheck{
		Id:        -1,
		ServiceId: "some-service",
		Timestamp: time.Time{},
		Metadata: map[string]string{
			"error": "much bad",
		},
	}
	checks := []*storage.HealthCheck{
		hc,
	}
	cfg := conf.ServiceConfig{
		Timeout: 24 * time.Hour,
	}

	status := monitor.GetServiceStatus(cfg, checks)

	require.Equal(t, monitor.STATUS_FAIL, status)
}
