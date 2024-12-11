package cmd

import (
	"log"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/status"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

		db, err := storage.InitDB()
		if err != nil {
			log.Fatal("Failed to initialize database:", err)
			return err
		}
		defer db.Close()

		heartbeatServer := monitor.HeartbeatListener{}
		config := viper.New()
		config.Set("port", port)
		heartbeatServer.Start(db, config)
		config.Set("port", guiPort)
		status.StartWebUI(db, config)
		exit := make(chan struct{})
		<-exit
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().Int("port", 8088, "Port where the heartbeat server should run")
	startCmd.Flags().Int("gui-port", 8089, "Port where the GUI server should run")
}
