package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

//go:embed templates/*
var TEMPLATES embed.FS

const (
	SUMMARY_STATS_LOOKBACK = -30 * 24 * time.Hour
	SUMMARY_STATS_INTERVAL = 30 * time.Minute
)

// Show services status
func handleIndex(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.Get()
		// Prepare a map to hold services and their heartbeats
		type ServiceStatus struct {
			// Needed as HealthCheck can be nil
			ServiceId         string
			LatestHealthCheck *storage.HealthCheck
			CurrentStatus     monitor.ServiceStatus
			UptimeSummary     string
		}
		var services []ServiceStatus
		serviceChecker := DefaultServiceChecker()

		now := time.Now().UTC()
		from := now.Add(SUMMARY_STATS_LOOKBACK)

		for _, serviceCfg := range config.AllServices() {
			logger.Debugw("Querying", "service", serviceCfg.Id)
			checks, err := db.HealthChecksSince(serviceCfg.Id, from)
			if err != nil {
				logger.Errorw("Failed to load health checks", "service", serviceCfg.Id, zap.Error(err))
				http.Error(w, "Failed to load health checks", http.StatusInternalServerError)
				return
			}

			var latestCheck *storage.HealthCheck
			if len(checks) > 0 {
				latestCheck = checks[len(checks)-1]
			}

			serviceStatus := serviceChecker.GetServiceStatus(latestCheck)

			intervals := monitor.BuildStatusIntervals(checks, from, now, SUMMARY_STATS_INTERVAL)
			up, down := monitor.SummarizeIntervals(intervals)
			uptimeSummary := fmt.Sprintf("%.2f%% up, %.2f%% down", up, down)

			// Add service and its heartbeats to the list
			services = append(services, ServiceStatus{
				ServiceId:         serviceCfg.Id,
				LatestHealthCheck: latestCheck,
				UptimeSummary:     uptimeSummary,
				CurrentStatus:     serviceStatus,
			})
		}

		hasMetadata := func(hc *storage.HealthCheck) bool {
			return hc != nil && len(hc.Metadata) > 0
		}

		timeAgoHealthCheck := func(hc *storage.HealthCheck) string {
			if hc == nil {
				return "never"
			}
			return TimeAgo(hc.Timestamp)
		}

		funcMap := template.FuncMap{
			"TimeAgo":     timeAgoHealthCheck,
			"HasMetadata": hasMetadata,
		}

		tmpl := template.New("index.html").Funcs(funcMap)
		// todo: might want to parse this only once
		tmpl, err := tmpl.ParseFS(TEMPLATES,
			filepath.Join("templates", "index.html"),
			filepath.Join("templates", "header.html"),
			filepath.Join("templates", "common.css"),
		)
		if err != nil {
			logger.Errorw("Error parsing template", zap.Error(err))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, map[string]any{
			"services":    services,
			"CurrentPage": "home",
		})
		if err != nil {
			logger.Errorw("Error rendering template", zap.Error(err))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}

// Show services status
func handleAbout(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.Get()
		tmpl := template.New("about.html")
		tmpl, err := tmpl.ParseFS(TEMPLATES,
			filepath.Join("templates", "about.html"),
			filepath.Join("templates", "header.html"),
			filepath.Join("templates", "common.css"),
		)
		if err != nil {
			logger.Errorw("Error parsing template", zap.Error(err))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}

		timeFormat := "15:04 Monday 02 January"
		lastReport, err := db.LatestTaskLog("report")
		if err != nil {
			logger.Errorw("DB error", zap.Error(err))
			http.Error(w, "Server error, please try again later", http.StatusInternalServerError)
			return
		}

		buildInfo, ok := debug.ReadBuildInfo()
		var beaconVersion string
		// manually specified
		if conf.GitRef != "" {
			beaconVersion = fmt.Sprintf("%s-%s", conf.GitRef, conf.GitSha)
		} else if ok { // auto-detect from .git
			beaconVersion = buildInfo.Main.Version
		} else {
			beaconVersion = "unknown"
		}

		var lastReportTime, nextReportAfter string
		lastReportStatus := ""
		if lastReport == nil {
			lastReportTime = "never"
			nextReportAfter = "error"
			logger.Error("DB not properly initialized! No report task found")
		} else if lastReport.Status == string(TASK_SENTINEL) {
			lastReportTime = "never"
			nextReportAfter = NextReportTime(config, lastReport.Timestamp).
				In(config.Timezone.Location).Format(timeFormat)
		} else {
			lastReportTime = lastReport.Timestamp.In(config.Timezone.Location).Format(timeFormat)
			lastReportStatus = lastReport.Status
			nextReportAfter = NextReportTime(config, lastReport.Timestamp).
				In(config.Timezone.Location).Format(timeFormat)
		}
		if !config.EmailConf.IsEnabled() {
			// todo: more info would be nice (why disabled)
			nextReportAfter = "disabled"
		}
		serverTime := time.Now().In(config.Timezone.Location).Format(timeFormat)

		zone, offset := time.Now().In(config.Timezone.Location).Zone()

		err = tmpl.Execute(w, map[string]any{
			"lastReportTime":        lastReportTime,
			"lastReportStatus":      lastReportStatus,
			"serverTime":            serverTime,
			"nextReportTime":        nextReportAfter,
			"ReportAfter":           config.ReportAfter,
			"CurrentPage":           "about",
			"Timezone":              config.Timezone.Location.String(),
			"TimezoneAlt":           zone,
			"TimezoneOffsetMinutes": offset / 60,
			"BeaconVersion":         beaconVersion,
		})
		if err != nil {
			logger.Errorw("Failed to render", zap.Error(err))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}
