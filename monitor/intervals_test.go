package monitor

import (
	"testing"
	"time"

	"github.com/davidmasek/beacon/logging"
	"github.com/stretchr/testify/assert"
)

func TestDurationHuman(t *testing.T) {
	logging.InitTest(t)

	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		interval Interval
		expected string
	}{
		{
			name: "7 days duration",
			interval: Interval{
				Start: fixedTime,
				End:   fixedTime.Add(7 * 24 * time.Hour),
			},
			expected: "7 days",
		},
		{
			name: "1 hour duration",
			interval: Interval{
				Start: fixedTime,
				End:   fixedTime.Add(1 * time.Hour),
			},
			expected: "1 hour",
		},
		{
			name: "zero duration",
			interval: Interval{
				Start: fixedTime,
				End:   fixedTime,
			},
			expected: "< 1 minute",
		},
		{
			name: "few minutes",
			interval: Interval{
				Start: fixedTime,
				End:   fixedTime.Add(7 * time.Minute),
			},
			expected: "7 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.interval.DurationHuman()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
