package monitor

import (
	"time"

	"github.com/davidmasek/beacon/storage"
)

type IntervalStatus struct {
	Interval Interval
	Status   ServiceStatus
}

func BuildStatusIntervals(checks []*storage.HealthCheck, from, to time.Time, interval time.Duration) []IntervalStatus {
	intervals := []IntervalStatus{}
	checkIdx := 0

	for ts := from; ts.Before(to); ts = ts.Add(interval) {
		currentEnd := ts.Add(interval)
		// Find the latest check at or before this interval end
		var latest *storage.HealthCheck
		for checkIdx < len(checks) && !checks[checkIdx].Timestamp.After(currentEnd) {
			latest = checks[checkIdx]
			checkIdx++
		}

		// Decide if latest is recent enough and OK
		status := STATUS_FAIL
		if latest != nil &&
			// not older than 1 interval
			latest.Timestamp.After(ts.Add(-interval)) {
			status = HealthCheckStatus(latest)
		}

		intervals = append(intervals, IntervalStatus{
			Interval: Interval{Start: ts, End: currentEnd},
			Status:   status,
		})
	}

	return intervals
}

func SummarizeIntervals(intervals []IntervalStatus) (upPct, downPct float64) {
	if len(intervals) == 0 {
		return 0, 0
	}

	var up, down int
	for _, interval := range intervals {
		switch interval.Status {
		case "OK":
			up++
		default:
			down++
		}
	}

	total := up + down
	upPct = float64(up) * 100 / float64(total)
	downPct = 100 - upPct
	return
}
