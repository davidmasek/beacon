package reporting

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/scheduler"
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

func TestRunAllJobs(t *testing.T) {
	logger := logging.InitTest(t)

	db := storage.NewTestDb(t)
	defer db.Close()

	beaconGithub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Beacon Github Page"))
		if err != nil {
			logger.Error("Failed to write response form BeaconGithub test server")
		}
	}))
	defer beaconGithub.Close()
	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer tsFail.Close()
	tsDisabled := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Error("Disabled server called - should not be called")
		t.Fail()
	}))
	defer tsDisabled.Close()
	configTemplate := `
services:
  beacon-github:
    url: "%s"
    status:
      - 200
    content:
      - Beacon
  beacon-periodic-checker:
  example-basic-web:
    url: "%s"
  example-temp-disable:
    url: "%s"
    enabled: false`
	configStr := fmt.Sprintf(configTemplate, beaconGithub.URL, tsFail.URL, tsDisabled.URL)

	config, err := conf.ConfigFromBytes([]byte(configStr))
	require.NoError(t, err)
	// disable emails for the test
	config.EmailConf.Enabled = "false"

	tmp, err := os.CreateTemp("", "beacon-test-report-*.html")
	require.NoError(t, err)
	tmp_file := tmp.Name()
	defer os.Remove(tmp_file)
	t.Logf("Using tmp file: %q\n", tmp_file)

	config.ReportName = tmp_file
	err = RunAllJobs(db, config, time.Now())
	require.NoError(t, err)
	require.FileExists(t, tmp_file)

	dat, err := os.ReadFile(tmp_file)
	require.NoError(t, err)
	require.NotEmpty(t, dat, "Report is empty")
	t.Log(string(dat))
	content := string(dat)
	require.Contains(t, content, "<html")
	require.Contains(t, content, "beacon-github")
	require.Contains(t, content, "beacon-periodic-checker")
	require.Contains(t, content, "example-basic-web")
	require.Contains(t, content, "example-temp-disable")
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
			return report.ServiceCfg.Id == serviceId
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
	next := scheduler.NextReportTime(config, now).In(timezone)
	assert.Equal(t, 10, next.Hour(), next.In(time.UTC))
	assert.Equal(t, now.In(timezone).Day()+1, next.Day(), fmt.Sprintf("now: %s, next: %s\n", now, next))
}

func TestNextReportTimeWeekends(t *testing.T) {
	config := conf.NewConfig()
	config.ReportAfter = 10
	config.Timezone = conf.TzLocation{Location: time.UTC}
	monday := time.Date(2025, 4, 28, 12, 0, 0, 0, time.UTC)

	next := scheduler.NextReportTime(config, monday)
	require.Equal(t, "Tuesday", next.Weekday().String())

	weekdays := conf.WeekdaysSet{}
	err := weekdays.ParseString("Wed")
	require.NoError(t, err)
	config.ReportOnDays = weekdays
	next = scheduler.NextReportTime(config, monday)
	require.Equal(t, "Wednesday", next.Weekday().String())

	err = weekdays.ParseString("Sun Mon")
	require.NoError(t, err)
	config.ReportOnDays = weekdays
	next = scheduler.NextReportTime(config, monday)
	require.Equal(t, "Sunday", next.Weekday().String())

	err = weekdays.ParseString("Mon")
	require.NoError(t, err)
	config.ReportOnDays = weekdays
	next = scheduler.NextReportTime(config, monday)
	require.Equal(t, "Monday", next.Weekday().String())
}
