package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestEndToEndHeartbeatCLI(t *testing.T) {
	// https://pkg.go.dev/os/exec#CommandContext might be of interest for future improvements
	// such as timeouts
	setupDbPathEnv(t)

	// list services
	cmd := exec.Command("go", "run", "..", "list")
	t.Log(cmd)
	output, err := cmd.CombinedOutput()
	require.Nil(t, err, output)

	// verify no services known
	// we should be running with empty DB
	// if there is something already - abort
	require.Empty(t, output)

	heartbeatPort := "9100"
	uiPort := "9101"
	// needs separate variable so we keep the reference to cleanup
	serverCmd := exec.Command("go", "run", "..", "start", "--port", heartbeatPort, "--gui-port", uiPort)
	t.Log(serverCmd)
	err = serverCmd.Start()
	defer func() {
		// TODO: this still does not properly cleanup sometimes
		err := serverCmd.Process.Kill()
		t.Log("exit... Err:", err)
		err = serverCmd.Wait()
		t.Log("done. Err:", err)

	}()
	require.Nil(t, err)

	service_name := "heartbeat-monitor"
	serverAddress := fmt.Sprint("http://localhost:", heartbeatPort)

	// wait for server to start
	time.Sleep(100 * time.Millisecond)

	t.Log("Record heartbeat")
	cmd = exec.Command("go", "run", "..", "heartbeat", service_name, "--server", serverAddress)
	t.Log(cmd)
	output, err = cmd.CombinedOutput()
	require.Nil(t, err, string(output))
	// example output: http://localhost:9000/beat/heartbeat-monitor 200 OK heartbeat-monitor @ 2024-12-11T22:46:44Z
	t.Log(string(output))
	require.Contains(t, string(output), " @ ")
	parts := strings.Split(string(output), " ")
	separatorIndex := 0
	for i, part := range parts {
		if part == "@" {
			separatorIndex = i
			break
		}
	}
	timestampIn := parts[separatorIndex+1]

	t.Log("Retrieve heartbeat status")
	cmd = exec.Command("go", "run", "..", "status", service_name, "--server", serverAddress)
	// example output: heartbeat-monitor @ 2024-12-11T22:46:44Z
	t.Log(cmd)
	output, err = cmd.CombinedOutput()
	require.Nil(t, err, output)
	t.Log(string(output))
	timestampOut := strings.Split(string(output), " ")[2]

	assert.Equal(t, timestampIn, timestampOut)

	t.Log("Check web UI")
	html := getHTML("/", t, uiPort)
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
