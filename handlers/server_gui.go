package handlers

import (
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

// Show services status
func handleIndex(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceNames, err := db.ListServices()
		if err != nil {
			http.Error(w, "Failed to fetch services", http.StatusInternalServerError)
			return
		}

		// Prepare a map to hold services and their heartbeats
		type ServiceStatus struct {
			// Needed as HealthCheck can be nil
			ServiceId         string
			LatestHealthCheck *storage.HealthCheck
			CurrentStatus     monitor.ServiceStatus
		}
		var services []ServiceStatus
		serviceChecker := DefaultServiceChecker()
		for _, serviceId := range serviceNames {
			healthCheck, err := db.LatestHealthCheck(serviceId)
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
				ServiceId:         serviceId,
				LatestHealthCheck: healthCheck,
				CurrentStatus:     serviceStatus,
			})
		}

		// Sort services alphabetically (optional, already done by SQL, kept for future use/changes)
		sort.Slice(services, func(i, j int) bool {
			return services[i].LatestHealthCheck.ServiceId < services[j].LatestHealthCheck.ServiceId
		})

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
		cwd, _ := os.Getwd()
		path := filepath.Join(cwd, "templates", "index.html")
		// workaround for tests than need different relative path
		// better fix wanted
		_, err = os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			path = filepath.Join(cwd, "..", "templates", "index.html")
		}
		tmpl, err = tmpl.ParseFiles(path)
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, services)
		if err != nil {
			log.Println("Failed to render", err)
		}
	}
}
