package monitor_test

import (
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
)

func mustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestBuildStatusIntervals(t *testing.T) {
	interval := 30 * time.Minute
	from := mustParse("2025-01-01T00:00:00Z")
	to := mustParse("2025-01-01T03:00:00Z") // 3 hours = 6 intervals

	tests := []struct {
		name         string
		checks       []*storage.HealthCheck
		expectStatus []string // expected status per interval
	}{
		{
			name:         "no checks = all FAIL",
			checks:       nil,
			expectStatus: []string{"FAIL", "FAIL", "FAIL", "FAIL", "FAIL", "FAIL"},
		},
		{
			name: "single OK at 00:00 = only first interval OK",
			checks: []*storage.HealthCheck{
				{Timestamp: mustParse("2025-01-01T00:00:00Z"), Metadata: map[string]string{"status": "OK"}},
			},
			expectStatus: []string{"OK", "FAIL", "FAIL", "FAIL", "FAIL", "FAIL"},
		},
		{
			name: "OK every ~30m = all OK",
			checks: []*storage.HealthCheck{
				{Timestamp: mustParse("2025-01-01T00:01:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T00:35:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T01:07:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T01:40:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T02:15:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T02:45:00Z"), Metadata: map[string]string{"status": "OK"}},
			},
			expectStatus: []string{"OK", "OK", "OK", "OK", "OK", "OK"},
		},
		{
			name: "OK at 00:00 and FAIL at 01:00",
			checks: []*storage.HealthCheck{
				{Timestamp: mustParse("2025-01-01T00:00:00Z"), Metadata: map[string]string{"status": "OK"}},
				{Timestamp: mustParse("2025-01-01T01:00:00Z"), Metadata: map[string]string{"status": "FAIL"}},
			},
			expectStatus: []string{"OK", "FAIL", "FAIL", "FAIL", "FAIL", "FAIL"},
		},
		{
			name: "OK at 00:45 = affects only 01:00 interval",
			checks: []*storage.HealthCheck{
				{Timestamp: mustParse("2025-01-01T00:45:00Z"), Metadata: map[string]string{"status": "OK"}},
			},
			expectStatus: []string{"FAIL", "OK", "FAIL", "FAIL", "FAIL", "FAIL"},
		},
		{
			name: "old OK at 23:00 = doesn't count",
			checks: []*storage.HealthCheck{
				{Timestamp: mustParse("2024-12-31T23:00:00Z"), Metadata: map[string]string{"status": "OK"}},
			},
			expectStatus: []string{"FAIL", "FAIL", "FAIL", "FAIL", "FAIL", "FAIL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intervals := monitor.BuildStatusIntervals(tt.checks, from, to, interval)

			var actual []string
			for _, iv := range intervals {
				actual = append(actual, string(iv.Status))
			}

			assert.Equal(t, tt.expectStatus, actual)
		})
	}
}
