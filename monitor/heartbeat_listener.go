package monitor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/davidmasek/beacon/storage"
)

type HeartbeatListener struct {
}

// Verify that HeartbeatListener implements Monitor interface
var _ Monitor = (*HeartbeatListener)(nil)

func (*HeartbeatListener) Start(db storage.Storage, viper *viper.Viper) error {
	http.HandleFunc("/beat/{service_id}", handleBeat(db))
	http.HandleFunc("/status/{service_id}", handleStatus(db))
	viper.SetDefault("port", "8088")
	port := viper.GetString("port")

	go func() {
		fmt.Printf("Starting HeartbeatListener server on http://localhost:%s\n", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
			log.Print(err)
			panic(err)
		}
	}()
	return nil
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
		fmt.Fprintf(w, "%s @ %s", serviceID, nowStr)
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
			http.Error(w, "No records for given service", http.StatusNotFound)
		}
		timestamp := timestamps[0]

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s @ %s", serviceID, timestamp.UTC().Format(storage.TIME_FORMAT))
	}
}
