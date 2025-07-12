package monitor

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
)

type HeartbeatResponse struct {
	ServiceId string `json:"service_id"`
	Timestamp string `json:"timestamp"`
}

type StatusResponse struct {
	ServiceId string `json:"service_id"`
	Timestamp string `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
}

func RegisterHeartbeatHandlers(db storage.Storage, mux *http.ServeMux, config *conf.Config) {
	mux.HandleFunc("/services/{service_id}/beat", handleBeat(db, config))
	mux.HandleFunc("/services/{service_id}/status", handleStatus(db, config))
}

// Auth might be better handled by middleware function instead,
// which might also extract and validate the service (or another middleware could do that).
// For now, going with this simple function that returns if processing should stop (auth failed or something else is wrong).
func checkAuth(w http.ResponseWriter, r *http.Request, config *conf.Config, service *conf.ServiceConfig) (stop bool) {
	logger := logging.Get()
	// service not found and unknown services not allowed
	if service == nil && !config.AllowUnknownHeartbeats {
		http.Error(w, "Service not found", http.StatusNotFound)
		return true
	}

	refToken := ""
	if service != nil {
		refToken = service.Token.Get()
	}
	// auth not configured for a given service, but required
	if refToken == "" && config.RequireHeartbeatAuth {
		logger.Warnw("Auth required but not configured for a service", "service", service.Id)
		http.Error(w, "Service auth required but not configured", http.StatusForbidden)
		return true
	}

	// auth not required
	if refToken == "" {
		return false
	}

	authHeader := r.Header.Get("Authorization")
	token := ""
	// auth not present
	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Debug("no auth")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return true
	}
	token = strings.TrimPrefix(authHeader, "Bearer ")

	// bad token
	if token != refToken {
		logger.Debug("bad auth")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return true
	}
	return false
}

func handleBeat(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.Get()
		serviceId := r.PathValue("service_id")
		if serviceId == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}
		service := config.Services.Get(serviceId)
		stop := checkAuth(w, r, config, service)
		if stop {
			return
		}

		now := time.Now()
		// Log the heartbeat to the database
		nowStr, err := db.RecordHeartbeat(serviceId, now)
		if err != nil {
			logger.Error("Failed to log heartbeat", zap.Error(err))
			http.Error(w, "Failed to log heartbeat", http.StatusInternalServerError)
			return
		}

		response := HeartbeatResponse{
			ServiceId: serviceId,
			Timestamp: nowStr,
		}

		// Respond to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.Error("Failed to encode /beat response", zap.Error(err))
		}
	}
}

func handleStatus(db storage.Storage, config *conf.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.Get()
		serviceId := r.PathValue("service_id")
		if serviceId == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}
		service := config.Services.Get(serviceId)
		stop := checkAuth(w, r, config, service)
		if stop {
			return
		}

		// Query the database for the latest heartbeat
		timestamps, err := db.GetLatestHeartbeats(serviceId, 1)
		if err != nil {
			logger.Error(err)
			http.Error(w, "Failed to query database", http.StatusInternalServerError)
			return
		}
		var response StatusResponse
		if len(timestamps) == 0 {
			response = StatusResponse{
				ServiceId: serviceId,
				Message:   "never",
			}
		} else {
			response = StatusResponse{
				ServiceId: serviceId,
				Timestamp: timestamps[0].UTC().Format(storage.TIME_FORMAT),
			}
		}

		// Respond to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.Error("Failed to encode /status response", zap.Error(err))
		}
	}
}
