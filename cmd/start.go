package cmd

import (
	"context"
	"fmt"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/reporting"
	"github.com/davidmasek/beacon/storage"
	"github.com/davidmasek/beacon/web_server"
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

		logger := logging.Get()
		logger.Info(">>> Beacon startup <<<")
		logger.Info("^^^^^^^^^^^^^^^^^^^^^^")

		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		db, err := storage.InitDB(config.DbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Overwrite existing config only if set on CLI.
		// Prefer existing config otherwise, ignore the "Cobra" default.
		portSet := cmd.Flag("port").Changed
		if portSet {
			config.Port = port
		}
		server, err := web_server.StartServer(db, config)
		if err != nil {
			return err
		}

		if !config.EmailConf.IsConfigured() {
			missingFields := config.EmailConf.MissingConfigurationFields()
			logger.Warnw("Email not configured", "missing fields", missingFields)
		}

		ctx, cancelScheduler := context.WithCancel(context.Background())
		go reporting.Start(ctx, db, config)

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

	startCmd.Flags().Int("port", 0, "Port where the server should run")
	startCmd.Flags().Bool("stop", false, "Stop the server after starting")
}
