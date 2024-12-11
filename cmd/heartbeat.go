package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat service-id",
	Args:  cobra.ExactArgs(1),
	Short: "Send heartbeat for a service to Beacon",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			return err
		}
		return sendHeartbeat(server, args[0])
	},
}

func init() {
	rootCmd.AddCommand(heartbeatCmd)
}

func sendHeartbeat(server string, serviceId string) error {
	target := fmt.Sprintf("%s/beat/%s", server, serviceId)
	resp, err := http.Post(target, "application/json", nil)
	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		} else {
			fmt.Printf(
				"%s %s %s\n",
				target,
				resp.Status,
				string(body),
			)
		}
	}
	return nil
}
