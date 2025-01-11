package monitor

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

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

func RegisterHeartbeatHandlers(db storage.Storage, mux *http.ServeMux) {
	mux.HandleFunc("/services/{service_id}/beat", handleBeat(db))
	mux.HandleFunc("/services/{service_id}/status", handleStatus(db))
}

func handleBeat(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := r.PathValue("service_id")
		if serviceId == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}
		now := time.Now()
		// Log the heartbeat to the database
		nowStr, err := db.RecordHeartbeat(serviceId, now)
		if err != nil {
			log.Println("[ERROR]", err)
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
		json.NewEncoder(w).Encode(response)
	}
}

func handleStatus(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := r.PathValue("service_id")
		if serviceId == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}

		// Query the database for the latest heartbeat
		timestamps, err := db.GetLatestHeartbeats(serviceId, 1)
		if err != nil {
			log.Println("[ERROR]", err)
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
		json.NewEncoder(w).Encode(response)
	}
}
