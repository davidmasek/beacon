package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

func StartServer(db storage.Storage, config *conf.Config) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", handleIndex(db, config))
	mux.HandleFunc("/about", handleAbout(db, config))

	monitor.RegisterHeartbeatHandlers(db, mux)
	port := config.Port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		fmt.Printf("Starting UI server on http://localhost:%d\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Print(err)
			panic(err)
		}
	}()
	return server, nil
}
