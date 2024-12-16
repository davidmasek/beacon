package handlers

import (
	"slices"
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Prepare test DB
func setupDB(t *testing.T) storage.Storage {
	db, err := storage.NewSQLStorage(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// TODO/feature: FIXME
func TestWriteReport(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	serviceGood := "alpha-should-pass"
	serviceTimeout := "beta-should-fail"
	serviceStatusFail := "gamma-should-fail"
	serviceError := "delta-should-fail"
	serviceGoodWithStatus := "epsilon-should-pass"
	expectedStates := map[string]monitor.ServiceState{
		serviceGood:           monitor.STATUS_OK,
		serviceTimeout:        monitor.STATUS_FAIL,
		serviceStatusFail:     monitor.STATUS_FAIL,
		serviceError:          monitor.STATUS_FAIL,
		serviceGoodWithStatus: monitor.STATUS_OK,
	}

	err := db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: serviceGood,
		Timestamp: time.Now().Add(-time.Hour),
	})
	require.NoError(t, err)
	err = db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: serviceTimeout,
		Timestamp: time.Now().Add(-time.Hour * 24 * 14),
	})
	require.NoError(t, err)
	err = db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: serviceStatusFail,
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"status": "totally-not-good",
		},
	})
	require.NoError(t, err)
	err = db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: serviceError,
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"status": string(monitor.STATUS_OK),
		},
	})
	require.NoError(t, err)
	err = db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: serviceGoodWithStatus,
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"error": "something really bad happened",
		},
	})
	require.NoError(t, err)

	reports, err := GenerateReport(db)
	require.NoError(t, err)

	assert.Len(t, reports, len(expectedStates))
	for serviceId := range expectedStates {
		idx := slices.IndexFunc(reports, func(report ServiceReport) bool {
			return report.ServiceId == serviceId
		})
		require.GreaterOrEqualf(t, idx, 0, "Service %s not found in reports", serviceId)

		reported := reports[idx].ServiceStatus
		expected := expectedStates[serviceId]
		assert.Equal(t, expected, reported)
	}
}
