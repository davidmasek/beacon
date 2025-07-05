package monitor

import (
	"fmt"
	"time"
)

type Interval struct {
	Start time.Time
	End   time.Time
}

func (interval *Interval) Duration() time.Duration {
	return interval.End.Sub(interval.Start)
}

func (interval *Interval) DurationHuman() string {
	duration := interval.Duration()
	switch {
	case duration.Hours() > 24:
		return fmt.Sprintf("%d days", int(duration.Hours()/24))
	case duration.Hours() == 1:
		return fmt.Sprintf("%d hour", int(duration.Hours()))
	case duration.Hours() >= 1:
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	case duration.Minutes() == 1:
		return fmt.Sprintf("%d minute", int(duration.Minutes()))
	case duration.Minutes() >= 1:
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	default:
		return "< 1 minute"
	}
}
