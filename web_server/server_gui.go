package web_server

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
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

const (
	SUMMARY_STATS_LOOKBACK = -30 * 24 * time.Hour
	SUMMARY_STATS_INTERVAL = 30 * time.Minute
)

var (
	//go:embed templates/*
	TEMPLATES      embed.FS
	INDEX_TEMPLATE *template.Template
	ABOUT_TEMPLATE *template.Template
)

func init() {
	funcMap := template.FuncMap{
		"HealthCheckStatus": monitor.HealthCheckStatus,
		"TimeAgo":           TimeAgo,
	}

	tmpl := template.New("index.html").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseFS(TEMPLATES,
		filepath.Join("templates", "index.html"),
		filepath.Join("templates", "header.html"),
		filepath.Join("templates", "common.css"),
	))

	INDEX_TEMPLATE = tmpl

	tmpl = template.New("about.html").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseFS(TEMPLATES,
		filepath.Join("templates", "about.html"),
		filepath.Join("templates", "header.html"),
		filepath.Join("templates", "common.css"),
	))

	ABOUT_TEMPLATE = tmpl
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

// Show services status
func handleIndex(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.Get()

		type ServiceView struct {
			ServiceId     string
			LastChecked   string
			CurrentStatus monitor.ServiceStatus
			UptimeSummary string
			RecentChecks  []*storage.HealthCheck
		}
		var services []ServiceView

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
			serviceStatus := monitor.GetServiceStatus(serviceCfg, checks)

			// Get the last checks for the details panel
			recentChecks, err := db.LatestHealthChecks(serviceCfg.Id, 5)
			if err != nil {
				logger.Errorw("Failed to load recent health checks", "service", serviceCfg.Id, zap.Error(err))
				http.Error(w, "Failed to load recent health checks", http.StatusInternalServerError)
				return
			}

			intervals := monitor.BuildStatusIntervals(checks, from, now, SUMMARY_STATS_INTERVAL)
			up, down := monitor.SummarizeIntervals(intervals)
			uptimeSummary := fmt.Sprintf("%.2f%% up, %.2f%% down", up, down)

			lastChecked := "never"
			if len(checks) > 0 {
				lastChecked = TimeAgo(checks[len(checks)-1].Timestamp)
			}

			services = append(services, ServiceView{
				ServiceId:     serviceCfg.Id,
				LastChecked:   lastChecked,
				UptimeSummary: uptimeSummary,
				CurrentStatus: serviceStatus,
				RecentChecks:  recentChecks,
			})
		}

		emailMissingConfig := config.EmailConf.MissingConfigurationFields()

		err := INDEX_TEMPLATE.Execute(w, map[string]any{
			"services":           services,
			"CurrentPage":        "home",
			"EmailMissingConfig": emailMissingConfig,
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
		} else if lastReport.Status == string(storage.TASK_SENTINEL) {
			lastReportTime = "never"
			nextReportAfter = scheduler.NextReportTime(config, lastReport.Timestamp).
				In(config.Timezone.Location).Format(timeFormat)
		} else {
			lastReportTime = lastReport.Timestamp.In(config.Timezone.Location).Format(timeFormat)
			lastReportStatus = lastReport.Status
			nextReportAfter = scheduler.NextReportTime(config, lastReport.Timestamp).
				In(config.Timezone.Location).Format(timeFormat)
		}
		if !config.EmailConf.IsEnabled() {
			if config.EmailConf.IsConfigured() {
				nextReportAfter = "disabled"
			} else {
				nextReportAfter = "disabled (emails not configured)"
			}
		}
		serverTime := time.Now().In(config.Timezone.Location).Format(timeFormat)

		zone, offset := time.Now().In(config.Timezone.Location).Zone()

		err = ABOUT_TEMPLATE.Execute(w, map[string]any{
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
