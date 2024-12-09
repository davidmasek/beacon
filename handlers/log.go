package handlers

import (
	"fmt"
	"log"

	"optimisticotter.me/heartbeat-monitor/monitor"
)

const (
	RESET  = "\033[0m"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
)

type LogHandler struct{}

func (lh LogHandler) Name() string {
	return "LogHandler"
}

func (lh LogHandler) Handle(reports []ServiceReport) error {
	for _, report := range reports {
		color := ""
		switch report.ServiceStatus {
		case monitor.STATUS_OK:
			color = GREEN
		case monitor.STATUS_OTHER:
			color = YELLOW
		case monitor.STATUS_FAIL:
			color = RED
		}
		statusWithWhitespace := fmt.Sprintf("%-5s", report.ServiceStatus)
		statusWithColors := fmt.Sprintf("%s%s%s", color, statusWithWhitespace, RESET)
		extra := ""
		if report.LatestHealthCheck != nil {
			extra = prettyPrint(report.LatestHealthCheck.Metadata)
		}
		log.Printf("[%s] %s Service %q. Details: %s", lh.Name(), statusWithColors, report.ServiceId, extra)
	}
	return nil
}
