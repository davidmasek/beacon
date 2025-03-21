package cmd

import (
	"fmt"

	"github.com/davidmasek/beacon/logging"
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
			logger.Errorw("Error message", zap.Error(fmt.Errorf("big bad")))
		}
		cmd.Println("Done")
		return nil
	},
}

func init() {
	debugCmd.Flags().Int("repeat", 1, "")

	rootCmd.AddCommand(debugCmd)
}
