package tests

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/davidmasek/beacon/handlers"
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

	config := viper.New()
	// shouldn't be fixed, but at least it's different than the default
	serverPort := "9000"
	config.Set("port", serverPort)
	t.Logf("Starting server on port %s\n", serverPort)
	server, err := handlers.StartServer(db, config)
	require.NoError(t, err)
	defer server.Close()

	// Is the sleep needed? Seems to work fine without
	// TODO: sometimes needed ... retry for Post might be nicer?
	time.Sleep(100 * time.Millisecond)

	t.Log("Record heartbeat")
	input := Post(fmt.Sprintf("/services/%s/beat", service_name), t, serverPort)
	assert.Contains(t, input, service_name)
	timestampIn := strings.Split(input, " ")[2]

	t.Log("Retrieve heartbeat status")
	output := Get(fmt.Sprintf("/services/%s/status", service_name), t, serverPort)
	assert.Contains(t, output, service_name)
	timestampOut := strings.Split(output, " ")[2]
	assert.Equal(t, timestampIn, timestampOut)

	t.Log("Check web UI")
	html := Get("/", t, serverPort)
	assert.Contains(t, html, "<html")
	assert.Contains(t, html, service_name)
}

// TODO: could replace with resty ?
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

// TODO: could replace with resty ?
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
