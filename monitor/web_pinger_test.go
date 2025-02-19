package monitor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckWebsite(t *testing.T) {
	logging.InitTest(t)
	tests := []struct {
		name         string
		httpStatus   []int
		bodyContent  []string
		responseCode int
		responseBody string
		expectStatus monitor.ServiceStatus
		expectErr    bool
	}{
		{
			name:         "Success - status matches and required content present",
			httpStatus:   []int{200, 201},
			bodyContent:  []string{"Hello", "Test"},
			responseCode: 200,
			responseBody: "Hello! This is a Test response.",
			expectStatus: monitor.STATUS_OK,
			expectErr:    false,
		},
		{
			name:         "Fail - status code not in expected list",
			httpStatus:   []int{200},
			bodyContent:  []string{"Ok"},
			responseCode: 404,
			responseBody: "Not found",
			expectStatus: monitor.STATUS_FAIL,
			expectErr:    false, // Because we return STATUS_FAIL with nil err for unexpected status code
		},
		{
			name:         "Fail - missing one of the required substrings in body",
			httpStatus:   []int{200},
			bodyContent:  []string{"Hello", "World"},
			responseCode: 200,
			responseBody: "Hello there but not the rest",
			expectStatus: monitor.STATUS_FAIL,
			expectErr:    false, // Missing content => STATUS_FAIL, nil error
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP server that returns a given status code and body.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Build your config
			config := &monitor.WebConfig{
				Url:         server.URL,
				HttpStatus:  tt.httpStatus,
				BodyContent: tt.bodyContent,
			}

			// Call the method under test
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			gotStatus, gotErr := config.CheckWebsite(ctx)

			// Validate the results
			assert.Equal(t, tt.expectStatus, gotStatus, "Unexpected service status.")
			if tt.expectErr {
				assert.Error(t, gotErr, "Expected an error but got nil.")
			} else {
				assert.NoError(t, gotErr, "Did not expect an error but got one.")
			}
		})
	}
}

func TestCheckWebsite_RequestError(t *testing.T) {
	logging.InitTest(t)
	// Provide an invalid URL to force http.Get to fail immediately.
	config := &monitor.WebConfig{
		Url: "http://invalid.invalid:9999", // Should fail DNS/connection
		// The following values won't matter because the request fails
		HttpStatus:  []int{200},
		BodyContent: []string{"irrelevant"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	gotStatus, gotErr := config.CheckWebsite(ctx)
	assert.Equal(t, monitor.STATUS_FAIL, gotStatus, "Expected STATUS_FAIL when request cannot be made.")
	assert.Error(t, gotErr, "Expected a non-nil error for a failing request.")
}

func TestCheckWebsite_BodyReadError(t *testing.T) {
	logging.InitTest(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content length, but never provide the content - should case "unexpected EOF"
		w.Header().Add("Content-Length", "2")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &monitor.WebConfig{
		Url:         ts.URL,
		HttpStatus:  []int{200},
		BodyContent: []string{"anything"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	gotStatus, gotErr := config.CheckWebsite(ctx)
	assert.Equal(t, monitor.STATUS_FAIL, gotStatus, "Expected STATUS_FAIL if body cannot be read.")
	assert.Error(t, gotErr, "Expected a read error.")
}

func TestCheckWebsite_Timeout(t *testing.T) {
	logging.InitTest(t)
	// Start a server that doesn't send a response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, err := w.(http.Hijacker).Hijack()
		require.NoError(t, err)
	}))
	defer ts.Close()

	config := &monitor.WebConfig{
		Url:        ts.URL,
		HttpStatus: []int{200},
	}

	// alternatively, one could remove the timeout and directly cancel the context
	// that would lead to faster runtime, but not sure if it would correctly
	// test the behavior
	timeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		Status monitor.ServiceStatus
		Err    error
	}
	done := make(chan result)

	start := time.Now()
	// Check async to prevent hanging here forever
	go func() {
		status, err := config.CheckWebsite(ctx)
		done <- result{status, err}
	}()

	var res result
	select {
	case res = <-done:
	case <-time.After(2 * timeout):
		elapsed := time.Since(start)
		t.Fatalf("CheckWebsite took too long, elapsed: %v", elapsed)
	}

	assert.Equal(t, monitor.STATUS_FAIL, res.Status, "Expected STATUS_FAIL for a timeout")
	assert.Error(t, res.Err, "Expected an error due to timeout")

}
