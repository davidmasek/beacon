package status

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

type HeartbeatConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

// Compute service status based on latest HealthCheck
// TODO: move this somewhere else
//
// TODO: maybe should be used like HealthCheck.GetStatus(config) or smth -> but this way
// enables healthcheck to be null without special handling outside which might be nice
func (config *HeartbeatConfig) GetServiceStatus(latestHealthCheck *storage.HealthCheck) (monitor.ServiceState, error) {
	if latestHealthCheck == nil {
		return monitor.STATUS_FAIL, nil
	}
	if time.Since(latestHealthCheck.Timestamp) > config.Timeout {
		return monitor.STATUS_FAIL, nil
	}

	if errorMeta, exists := latestHealthCheck.Metadata["error"]; exists {
		if errorMeta != "" {
			return monitor.STATUS_FAIL, nil
		}
	}
	if statusMeta, exists := latestHealthCheck.Metadata["status"]; exists {
		if statusMeta != string(monitor.STATUS_OK) {
			return monitor.STATUS_FAIL, nil
		}
	}
	return monitor.STATUS_OK, nil
}

// Handler for / to show services and recent heartbeats
// Enhancement: SQL Groupby instead of this bullshit
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
			CurrentStatus     monitor.ServiceState
		}
		var services []ServiceStatus
		for _, serviceId := range serviceNames {
			healthCheck, err := db.LatestHealthCheck(serviceId)
			if err != nil {
				// TODO: should log error
				http.Error(w, "Failed to fetch heartbeats", http.StatusInternalServerError)
				return
			}

			// TODO: currently uses default config build right here and ignores any settings
			// should load settings in the way it's done (depending on how it's done when we get to this)
			// TODO - BUG: this uses heartbeat config - but the service doesn't have to of that type (i.e. it's a website)
			config := HeartbeatConfig{Timeout: 24 * time.Hour}
			serviceStatus, err := config.GetServiceStatus(healthCheck)
			if err != nil {
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

func StartWebUI(db storage.Storage, config *viper.Viper) error {
	if config == nil {
		config = viper.New()
	}
	config.SetDefault("port", "8089")
	http.HandleFunc("/{$}", handleIndex(db))
	port := config.GetString("port")

	go func() {
		fmt.Printf("Starting UI server on http://localhost:%s\n", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
			log.Print(err)
			panic(err)
		}
	}()
	return nil
}
