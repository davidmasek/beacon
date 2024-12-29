package cmd

import (
	"context"
	"fmt"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var SERVER_SUCCESS_MESSAGE = "[SUCCESS] Startup complete. Stopping."

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Beacon server that listens for heartbeats and provides web GUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		stopServer, err := cmd.Flags().GetBool("stop")
		if err != nil {
			return err
		}

		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		db, err := storage.InitDB(config.GetString("DB"))
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// TODO: this currently overwrites other options
		// that should happen only if specified
		// Is it possible to distinguish with Cobra?
		config.Set("port", port)
		server, err := handlers.StartServer(db, config)
		if err != nil {
			return err
		}

		ctx, cancelScheduler := context.WithCancel(context.Background())
		go scheduler.Start(ctx, db, config)

		if stopServer {
			server.Close()
			cancelScheduler()
			cmd.Println(SERVER_SUCCESS_MESSAGE)
			return nil
		}

		exit := make(chan struct{})
		// block forever
		<-exit
		// not really needed, but linters complain otherwise
		cancelScheduler()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().Int("port", 8088, "Port where the server should run")
	startCmd.Flags().Bool("stop", false, "Stop the server after starting")
	startCmd.Flags().Bool("background", false, "Run in the background")
}
