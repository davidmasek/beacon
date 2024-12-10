package cmd

import (
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Use heartbeat API",
}

func init() {
	rootCmd.AddCommand(heartbeatCmd)

	heartbeatCmd.PersistentFlags().String("server", "http://localhost:8088", "Address of the target heartbeat server.")
}
