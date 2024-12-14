package tests

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/status"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
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

func TestEndToEndHeartbeat(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	service_name := "heartbeat-monitor"

	viper := viper.New()
	// shouldn't be fixed, but at least it's different than the default
	heartbeatPort := "9000"
	viper.Set("port", heartbeatPort)
	t.Logf("Starting heartbeat listener on port %s\n", heartbeatPort)
	heartbeatListener := monitor.HeartbeatListener{}
	heartbeatServer, err := heartbeatListener.Start(db, viper)
	require.NoError(t, err)
	defer heartbeatServer.Close()

	uiPort := "9001"
	viper.Set("port", uiPort)
	t.Logf("Starting web UI on port %s\n", heartbeatPort)
	uiServer, err := status.StartWebUI(db, viper)
	require.NoError(t, err)
	defer uiServer.Close()
	// Is the sleep needed? Seems to work fine without
	// TODO: sometimes needed ... retry for Post might be nicer?
	time.Sleep(100 * time.Millisecond)

	t.Log("Record heartbeat")
	input := Post(fmt.Sprintf("/beat/%s", service_name), t, heartbeatPort)
	assert.Contains(t, input, service_name)
	timestampIn := strings.Split(input, " ")[2]

	t.Log("Retrieve heartbeat status")
	output := Get(fmt.Sprintf("/status/%s", service_name), t, heartbeatPort)
	assert.Contains(t, output, service_name)
	timestampOut := strings.Split(output, " ")[2]
	assert.Equal(t, timestampIn, timestampOut)

	t.Log("Check web UI")
	html := Get("/", t, uiPort)
	assert.Contains(t, html, "<html")
	assert.Contains(t, html, service_name)
}

// TODO: could replace with http.Client ?
func Post(suffix string, t *testing.T, port string) string {
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s%s", port, suffix), "application/json", nil)
	if err != nil {
		t.Fatalf("Unable to POST to %s: %+v", suffix, err)
	}
	if resp != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unable to read response body from %s: %+v", suffix, err)
			return ""
		} else {
			var bodyInfo string
			if len(body) > 100 {
				bodyInfo = string(body[:10]) + "..."
			} else {
				bodyInfo = string(body)
			}
			t.Logf(
				"[INFO] %s %s %s",
				suffix,
				resp.Status,
				bodyInfo,
			)
			return string(body)
		}
	}
	return ""
}

// TODO: could replace with http.Client ?
func Get(suffix string, t *testing.T, port string) string {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", port, suffix))
	if err != nil {
		t.Fatalf("Unable to GET to %s: %+v", suffix, err)
	}
	if resp != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unable to read response body from %s: %+v", suffix, err)
			return ""
		} else {
			var bodyInfo string
			if len(body) > 100 {
				bodyInfo = string(body[:10]) + "..."
			} else {
				bodyInfo = string(body)
			}
			t.Logf(
				"[INFO] %s %s %s",
				suffix,
				resp.Status,
				bodyInfo,
			)
			return string(body)
		}
	}
	return ""
}
