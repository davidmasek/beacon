package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

//go:embed templates/*
var TEMPLATES embed.FS

// Show services status
func handleIndex(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare a map to hold services and their heartbeats
		type ServiceStatus struct {
			// Needed as HealthCheck can be nil
			ServiceId         string
			LatestHealthCheck *storage.HealthCheck
			CurrentStatus     monitor.ServiceStatus
		}
		var services []ServiceStatus
		serviceChecker := DefaultServiceChecker()
		for _, serviceCfg := range config.Services() {
			log.Println("Querying", serviceCfg.Id)
			healthCheck, err := db.LatestHealthCheck(serviceCfg.Id)
			if err != nil {
				log.Printf("Failed to load health check: %s", err)
				http.Error(w, "Failed to load health check", http.StatusInternalServerError)
				return
			}

			serviceStatus, err := serviceChecker.GetServiceStatus(healthCheck)
			if err != nil {
				log.Printf("Failed to calculate service status: %s", err)
				http.Error(w, "Failed to calculate service status", http.StatusInternalServerError)
				return
			}

			// Add service and its heartbeats to the list
			services = append(services, ServiceStatus{
				ServiceId:         serviceCfg.Id,
				LatestHealthCheck: healthCheck,
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
		path := filepath.Join("templates", "index.html")
		tmpl, err := tmpl.ParseFS(TEMPLATES, path)
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, map[string]any{
			"services": services,
		})
		if err != nil {
			log.Println("Failed to render", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}

// Show services status
func handleAbout(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.New("about.html")
		path := filepath.Join("templates", "about.html")
		tmpl, err := tmpl.ParseFS(TEMPLATES, path)
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}

		timeFormat := "15:04 Monday 02 January"
		lastReportTime := "never"
		serverTime := time.Now().In(&config.Timezone).Format(timeFormat)
		nextReportAfter := "TODO"

		err = tmpl.Execute(w, map[string]any{
			"lastReportTime": lastReportTime,
			"serverTime":     serverTime,
			"nextReportTime": nextReportAfter,
		})
		if err != nil {
			log.Println("Failed to render", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}
