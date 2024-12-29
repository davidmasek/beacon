package monitor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/davidmasek/beacon/storage"
)

func RegisterHeartbeatHandlers(db storage.Storage, mux *http.ServeMux) {
	mux.HandleFunc("/beat/{service_id}", handleBeat(db))
	mux.HandleFunc("/status/{service_id}", handleStatus(db))
}

// Handler for /beat/{service_id}
func handleBeat(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID := r.PathValue("service_id")
		if serviceID == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}
		now := time.Now()
		// Log the heartbeat to the database
		nowStr, err := db.RecordHeartbeat(serviceID, now)
		if err != nil {
			log.Println("[ERROR]", err)
			http.Error(w, "Failed to log heartbeat", http.StatusInternalServerError)
			return
		}

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s @ %s\n", serviceID, nowStr)
	}
}

func handleStatus(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID := r.PathValue("service_id")
		if serviceID == "" {
			http.Error(w, "Missing service_id", http.StatusBadRequest)
			return
		}

		// Query the database for the latest heartbeat
		timestamps, err := db.GetLatestHeartbeats(serviceID, 1)
		if err != nil {
			log.Println("[ERROR]", err)
			http.Error(w, "Failed to query database", http.StatusInternalServerError)
			return
		}
		if len(timestamps) == 0 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s @ never\n", serviceID)
			return
		}
		timestamp := timestamps[0]

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s @ %s\n", serviceID, timestamp.UTC().Format(storage.TIME_FORMAT))
	}
}
