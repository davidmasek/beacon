package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

type ServiceReport struct {
	ServiceId         string
	ServiceStatus     monitor.ServiceState
	LatestHealthCheck *storage.HealthCheck
}

type StatusHandler interface {
	Name() string
	Handle(servicesInfos []ServiceReport) error
}

func prettyPrint(details map[string]string) string {
	jsonData, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal details map: %w", err))
	}
	return string(jsonData)
}
