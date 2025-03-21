package monitor_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TEST_CFG = []byte(`
services:
    banana:
    orange:
        token: "juiceM8"
`)

func TestHandleBeat_Success(t *testing.T) {
	logging.InitTest(t)
	db := storage.NewTestDb(t)
	mux := http.NewServeMux()
	config := conf.NewConfig()
	// this test assumes no auth and unknown services allowed
	require.True(t, config.AllowUnknownHeartbeats)
	require.False(t, config.RequireHeartbeatAuth)
	monitor.RegisterHeartbeatHandlers(db, mux, config)

	// Send a "beat" for service ID "test-service"
	req := httptest.NewRequest(http.MethodPost, "/services/test-service/beat", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var hbResp monitor.HeartbeatResponse
	resp := w.Body.Bytes()
	err := json.Unmarshal(resp, &hbResp)
	t.Logf("Response: %q\n", string(resp))
	require.NoError(t, err, "Failed to parse heartbeat JSON response")
	assert.Equal(t, "test-service", hbResp.ServiceId)
	assert.NotEmpty(t, hbResp.Timestamp)

	_, err = db.GetLatestHeartbeats("test-service", 1)
	assert.NoError(t, err, "Should be able to read heartbeats from DB")
}

func TestHandleBeat_WithToken(t *testing.T) {
	logging.InitTest(t)
	db := storage.NewTestDb(t)
	mux := http.NewServeMux()
	config, err := conf.ConfigFromBytes(TEST_CFG)
	require.NoError(t, err)
	monitor.RegisterHeartbeatHandlers(db, mux, config)
	serviceId := "orange"
	beatUrl := fmt.Sprintf("/services/%s/beat", serviceId)

	req := httptest.NewRequest(http.MethodPost, beatUrl, nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	hb, err := db.GetLatestHeartbeats(serviceId, 1)
	require.NoError(t, err, "Should be able to read heartbeats from DB")
	require.Len(t, hb, 0, "No heartbeat should be stored")

	req = httptest.NewRequest(http.MethodPost, "/services/orange/beat", nil)
	req.Header.Add("Authorization", "Bearer juiceM8")
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var hbResp monitor.HeartbeatResponse
	err = json.Unmarshal(w.Body.Bytes(), &hbResp)
	require.NoError(t, err, "Failed to parse heartbeat JSON response")
	assert.Equal(t, serviceId, hbResp.ServiceId)
	assert.NotEmpty(t, hbResp.Timestamp)

	hb, err = db.GetLatestHeartbeats(serviceId, 1)
	require.NoError(t, err, "Should be able to read heartbeats from DB")
	require.Len(t, hb, 1, "Heartbeat should be stored")
}

func TestHandleStatus_HasHeartbeat(t *testing.T) {
	db := storage.NewTestDb(t)
	mux := http.NewServeMux()
	config := conf.NewConfig()
	monitor.RegisterHeartbeatHandlers(db, mux, config)

	// First, record a heartbeat
	_, err := db.RecordHeartbeat("alive-service", time.Now())
	assert.NoError(t, err)

	// Now check status
	req := httptest.NewRequest(http.MethodGet, "/services/alive-service/status", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse JSON
	var statusResp monitor.StatusResponse
	err = json.Unmarshal(w.Body.Bytes(), &statusResp)
	assert.NoError(t, err)
	assert.Equal(t, "alive-service", statusResp.ServiceId)
	assert.NotEmpty(t, statusResp.Timestamp)
	assert.Empty(t, statusResp.Message, "Should not have a 'message' if Timestamp is present")
}

func TestHandleStatus_NoHeartbeat(t *testing.T) {
	db := storage.NewTestDb(t)
	mux := http.NewServeMux()
	config := conf.NewConfig()
	monitor.RegisterHeartbeatHandlers(db, mux, config)

	// Query status for a service that has no heartbeats
	req := httptest.NewRequest(http.MethodGet, "/services/ghost-service/status", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse JSON response
	var statusResp monitor.StatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &statusResp)
	assert.NoError(t, err)
	assert.Equal(t, "ghost-service", statusResp.ServiceId)
	assert.Empty(t, statusResp.Timestamp)
	assert.Equal(t, "never", statusResp.Message, "Expected 'never' when no heartbeats exist")
}
