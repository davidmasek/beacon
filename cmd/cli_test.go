package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDbPathEnv(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-beacon")
	require.Nil(t, err)
	t.Log("tmp dir:", dir)
	tmp_file := filepath.Join(dir, "test.db")
	err = os.Setenv("BEACON_DB", tmp_file)
	require.Nil(t, err)
}

func TestStartServer(t *testing.T) {
	// be sure to setup DB path before interacting with Beacon
	setupDbPathEnv(t)

	serverPort := "9100"
	var outputBuffer bytes.Buffer
	rootCmd.SetOut(&outputBuffer)
	rootCmd.SetErr(&outputBuffer)
	outputBuffer.Reset()
	rootCmd.SetArgs([]string{"start", "--port", serverPort, "--stop"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	output := outputBuffer.String()
	require.Contains(t, output, SERVER_SUCCESS_MESSAGE)
}

func TestHeartbeat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	serverPort := "9100"

	// be sure to setup DB path before interacting with Beacon
	setupDbPathEnv(t)

	var outputBuffer bytes.Buffer
	rootCmd.SetOut(&outputBuffer)
	rootCmd.SetErr(&outputBuffer)

	// list services
	rootCmd.SetArgs([]string{"list"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	// verify no services known
	// we should be running with empty DB
	// if there is something already - abort
	require.Empty(t, outputBuffer)

	// We run the server directly from code (not using CLI).
	// TODO: server startup needs unification (CLI vs this test vs other tests)
	db, err := storage.InitDB()
	require.NoError(t, err)
	config := viper.New()
	config.Set("port", serverPort)
	server, err := handlers.StartServer(db, config)
	require.NoError(t, err)
	defer server.Close()

	// Is the sleep needed? Seems to work fine without
	// TODO: sometimes needed ... retry for Post might be nicer?
	time.Sleep(100 * time.Millisecond)

	service_name := "heartbeat-monitor"
	serverAddress := fmt.Sprint("http://localhost:", serverPort)

	t.Log("Record heartbeat")
	outputBuffer.Reset()
	rootCmd.SetArgs([]string{"heartbeat", service_name, "--server", serverAddress})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// example output: http://localhost:9000/beat/heartbeat-monitor 200 OK heartbeat-monitor @ 2024-12-11T22:46:44Z
	output := outputBuffer.String()

	require.Contains(t, output, " @ ")
	parts := strings.Split(output, " ")
	separatorIndex := 0
	for i, part := range parts {
		if part == "@" {
			separatorIndex = i
			break
		}
	}
	timestampIn := parts[separatorIndex+1]

	t.Log("Retrieve heartbeat status")
	outputBuffer.Reset()
	rootCmd.SetArgs([]string{"status", service_name, "--server", serverAddress})
	err = rootCmd.Execute()
	require.NoError(t, err)
	// example output: heartbeat-monitor @ 2024-12-11T22:46:44Z
	output = outputBuffer.String()
	timestampOut := strings.Split(output, " ")[2]

	assert.Equal(t, timestampIn, timestampOut)

	t.Log("Check web UI")
	html := getHTML("/", t, serverPort)
	assert.Contains(t, html, "<html")
	assert.Contains(t, html, service_name)
}

func getHTML(suffix string, t *testing.T, uiPort string) string {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", uiPort, suffix))
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
