package web_server

import (
	"fmt"
	"net/http"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
)

func StartServer(db storage.Storage, config *conf.Config) (*http.Server, error) {
	logger := logging.Get()
	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", handleIndex(db, config))
	mux.HandleFunc("/about", handleAbout(db, config))

	monitor.RegisterHeartbeatHandlers(db, mux, config)
	port := config.Port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		logger.Infow("Starting server", "host", fmt.Sprint("http://localhost:", port))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Panic(err)
		}
	}()
	return server, nil
}
