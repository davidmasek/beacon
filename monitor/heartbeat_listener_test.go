package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
)

func TestHandleStatus_HasHeartbeat(t *testing.T) {
	db := storage.NewTestDb(t)
	mux := http.NewServeMux()
	monitor.RegisterHeartbeatHandlers(db, mux)

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
	monitor.RegisterHeartbeatHandlers(db, mux)

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
