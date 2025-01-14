package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/storage"
)

func TestHandleIndex(t *testing.T) {
	db := storage.NewTestDb(t)
	defer db.Close()
	config, err := conf.ExampleConfig()
	require.NoError(t, err)

	// Construct the handler
	handler := handleIndex(db, config)

	// Make a test HTTP request
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check status code
	require.Equal(t, http.StatusOK, rr.Code, "handler should return 200 OK")

	// Check the response body (contains HTML). You can do more detailed checks if desired.
	body := rr.Body.String()
	// title/menu
	require.Contains(t, body, "Beacon")
	require.Contains(t, body, "About")
	// services
	require.Contains(t, body, "beacon-github", "Response should include \"beacon-github\" in rendered HTML")
}

func TestHandleAbout(t *testing.T) {
	// Create a mock database
	db := storage.NewTestDb(t)
	defer db.Close()
	config, err := conf.ExampleConfig()
	require.NoError(t, err)

	// Construct the handler
	handler := handleAbout(db, config)

	// Make a test HTTP request
	req := httptest.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check status code
	require.Equal(t, http.StatusOK, rr.Code, "handler should return 200 OK")

	// Check the response body (contains HTML). You can do more detailed checks if desired.
	body := rr.Body.String()
	// title/menu
	require.Contains(t, body, "Beacon")
	require.Contains(t, body, "About")
	// info
	require.Contains(t, body, "https://github.com/davidmasek/beacon")
	// config
	require.Contains(t, body, "Timezone:")
	require.Contains(t, body, "Europe/Prague")
}
