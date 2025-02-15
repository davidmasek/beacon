package cmd

import (
	"github.com/davidmasek/beacon/logging"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use: "debug",
	// Args:  cobra.ExactArgs(0),
	Short: "Development tools, use with caution",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.Get()
		logger.Debugw("Debug message", "foo", 42)
		logger.Infow("Info message", "foo", 42)
		logger.Warnw("Warn message", "foo", 42)
		logger.Errorw("Error message", "foo", 42)
		cmd.Println("Done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
