package jobs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/require"
)

func setupServers(t *testing.T) (config *conf.Config, teardown func()) {
	logger := logging.InitTest(t)
	beaconGithub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Beacon Github Page"))
		if err != nil {
			logger.Error("Failed to write response form BeaconGithub test server")
		}
	}))
	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	tsDisabled := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Error("Disabled server called - should not be called")
		t.Fail()
	}))
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
	return config, func() {
		beaconGithub.Close()
		tsFail.Close()
		tsDisabled.Close()
	}
}

func TestRunAllJobs(t *testing.T) {
	_ = logging.InitTest(t)

	db := storage.NewTestDb(t)
	defer db.Close()

	config, serverTeardown := setupServers(t)
	defer serverTeardown()
	// disable emails for the test
	config.EmailConf.Enabled = "false"

	tmp, err := os.CreateTemp("", "beacon-test-report-*.html")
	require.NoError(t, err)
	tmp_file := tmp.Name()
	defer os.Remove(tmp_file)
	t.Logf("Using tmp file: %q\n", tmp_file)

	// pruning: setup
	now := time.Now()
	err = db.AddHealthCheck(&storage.HealthCheckInput{
		ServiceId: "very-old-service",
		Timestamp: now.AddDate(-5, 0, 0),
	})
	require.NoError(t, err)

	config.ReportName = tmp_file
	err = RunAllJobs(db, config, now)
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

	hc, err := db.LatestHealthCheck("very-old-service")
	require.NoError(t, err)
	require.Nil(t, hc, "Old service health check should be pruned")
}
