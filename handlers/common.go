package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

type ServiceReport struct {
	ServiceId         string
	ServiceStatus     monitor.ServiceStatus
	LatestHealthCheck *storage.HealthCheck
}

func prettyPrint(details map[string]string) string {
	jsonData, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal details map: %w", err))
	}
	return string(jsonData)
}

// Human-readable time difference (e.g., "5 minutes ago")
func TimeAgo(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration.Hours() > 24:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration.Hours() == 1:
		return fmt.Sprintf("%d hour ago", int(duration.Hours()))
	case duration.Hours() >= 1:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration.Minutes() == 1:
		return fmt.Sprintf("%d minute ago", int(duration.Minutes()))
	case duration.Minutes() >= 1:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	default:
		return "just now"
	}
}
