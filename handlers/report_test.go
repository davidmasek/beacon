package handlers

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfgInput = []byte(`
services:
  recent-beat-should-pass:
  long-ago-should-fail:
  with-bad-status-should-fail:
  with-error-should-fail:
  with-explicit-status-should-pass:
`)

var testServicesInput = []storage.HealthCheckInput{
	{
		ServiceId: "recent-beat-should-pass",
		Timestamp: time.Now().Add(-time.Hour),
	},
	{
		ServiceId: "long-ago-should-fail",
		Timestamp: time.Now().Add(-time.Hour * 24 * 14),
	},
	{
		ServiceId: "with-bad-status-should-fail",
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"status": "totally-not-good",
		},
	},
	{
		ServiceId: "with-error-should-fail",
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"error": "cannot foo the bar",
		},
	},
	{
		ServiceId: "with-explicit-status-should-pass",
		Timestamp: time.Now().Add(-time.Hour),
		Metadata: map[string]string{
			"status": string(monitor.STATUS_OK),
		},
	},
}

func expectedStatusFromName(t *testing.T, name string) monitor.ServiceStatus {
	if strings.HasSuffix(name, "should-pass") {
		return monitor.STATUS_OK
	}
	if strings.HasSuffix(name, "should-fail") {
		return monitor.STATUS_FAIL
	}
	t.Fatalf("invalid test service name %q - invalid suffix", name)
	return monitor.STATUS_FAIL
}

func TestWriteReport(t *testing.T) {
	db := storage.NewTestDb(t)
	defer db.Close()

	config, err := conf.ConfigFromBytes(cfgInput)
	require.NoError(t, err)

	for _, input := range testServicesInput {
		err := db.AddHealthCheck(&input)
		require.NoError(t, err)
	}

	reports, err := GenerateReport(db, config)
	require.NoError(t, err)

	assert.Len(t, reports, len(testServicesInput))
	for _, input := range testServicesInput {
		serviceId := input.ServiceId
		idx := slices.IndexFunc(reports, func(report ServiceReport) bool {
			return report.ServiceId == serviceId
		})
		require.GreaterOrEqualf(t, idx, 0, "Service %s not found in reports", serviceId)

		reported := reports[idx].ServiceStatus
		expected := expectedStatusFromName(t, serviceId)
		assert.Equal(t, expected, reported)
	}
}

func TestNextReportTime(t *testing.T) {
	config := conf.NewConfig()
	config.ReportAfter = 10
	timezone, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	config.Timezone = conf.TzLocation{Location: timezone}

	now := time.Now()
	next := NextReportTime(config, now).In(timezone)
	assert.Equal(t, 10, next.Hour(), next.In(time.UTC))
	assert.Equal(t, now.In(timezone).Day()+1, next.Day(), fmt.Sprintf("now: %s, next: %s\n", now, next))
}
