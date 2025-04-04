package cmd

import (
	"fmt"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var debugCmd = &cobra.Command{
	Use: "debug",
	// Args:  cobra.ExactArgs(0),
	Short: "Development tools, use with caution",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.Get()
		repeat, err := cmd.Flags().GetInt("repeat")
		if err != nil {
			return err
		}
		for range repeat {
			logger.Debugw("Debug message", "foo", 42)
			logger.Infow("Info message", "foo", 42)
			logger.Warnw("Warn message", "foo", 42)
		}
		fail, err := cmd.Flags().GetBool("fail")
		if err != nil {
			return err
		}
		if fail {
			logger.Errorw("Error message", zap.Error(fmt.Errorf("big bad")))
			return fmt.Errorf("failed as expected")
		}
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		db, err := storage.InitDB(config.DbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()
		cmd.Println("DB Schema:")
		schemas, err := db.ListSchemaVersions()
		if err != nil {
			return fmt.Errorf("failed to list schemas: %w", err)
		}
		for _, schema := range schemas {
			cmd.Printf("- %#v\n", schema)
		}
		cmd.Println("Done")
		return nil
	},
}

func init() {
	debugCmd.Flags().Int("repeat", 1, "")
	debugCmd.Flags().Bool("fail", false, "")

	rootCmd.AddCommand(debugCmd)
}
