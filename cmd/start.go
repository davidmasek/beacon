package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/scheduler"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		guiPort, err := cmd.Flags().GetInt("gui-port")
		if err != nil {
			return err
		}
		stopServer, err := cmd.Flags().GetBool("stop")
		if err != nil {
			return err
		}

		db, err := storage.InitDB()
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		heartbeatListener := monitor.HeartbeatListener{}
		// TODO: need to unify config loading in CLI at least a bit
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		var config *viper.Viper
		if configFile == "" {
			config, err = monitor.DefaultConfig()
			// TODO: quick fix to enable start when no config file specified
			// should either decide to always require config file (and provide a reasonable default)
			// or handle this like normal behavior (i.e. not just hack it here)
			if err != nil && strings.Contains(err.Error(), `Config File "beacon.yaml" Not Found in`) {
				log.Println(err)
				cmd.Println("No config file found")
				err = nil
			}
		} else {
			config, err = monitor.DefaultConfigFrom(configFile)
		}
		if err != nil {
			return err
		}

		config.Set("port", port)
		heartbeatServer, err := heartbeatListener.Start(db, config)
		if err != nil {
			return err
		}
		config.Set("port", guiPort)
		uiServer, err := handlers.StartWebUI(db, config)
		if err != nil {
			return err
		}

		ctx, cancelScheduler := context.WithCancel(context.Background())
		go scheduler.Start(ctx, db, config)

		if stopServer {
			uiServer.Close()
			heartbeatServer.Close()
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

	startCmd.Flags().Int("port", 8088, "Port where the heartbeat server should run")
	startCmd.Flags().Int("gui-port", 8089, "Port where the GUI server should run")
	startCmd.Flags().Bool("stop", false, "Stop the server after starting")
	startCmd.Flags().Bool("background", false, "Run in the background")
}
