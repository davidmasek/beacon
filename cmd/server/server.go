// Run HeartbeatListener and web GUI.
// Listen for HTTP heartbeats and store to DB.
// Start web GUI server.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"optimisticotter.me/heartbeat-monitor/monitor"
	"optimisticotter.me/heartbeat-monitor/status"
	"optimisticotter.me/heartbeat-monitor/storage"
)

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	heartbeat_server := monitor.HeartbeatListener{}
	heartbeat_server.Start(db, nil)
	status.StartWebUI(db)
	time.Sleep(100 * time.Millisecond)
	Get("/beat/heartbeat-monitor")
	Get("/status/heartbeat-monitor")
	exit := make(chan struct{})
	<-exit
}

func Get(suffix string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8088%s", suffix))
	if err != nil {
		log.Println("[ERROR]", err)
	}
	if resp != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ERROR] Unable to read response body", err)
		} else {
			log.Println(
				"[INFO]",
				suffix,
				resp.Status,
				string(body),
			)
		}
	}
}
