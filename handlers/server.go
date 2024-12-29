package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

func StartServer(db storage.Storage, config *viper.Viper) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", handleIndex(db))

	monitor.RegisterHeartbeatHandlers(db, mux)

	if config == nil {
		config = viper.New()
	}
	config.SetDefault("port", "8088")
	port := config.GetString("port")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	go func() {
		fmt.Printf("Starting UI server on http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Print(err)
			panic(err)
		}
	}()
	return server, nil
}
