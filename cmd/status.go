package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status service-id",
	Args:  cobra.ExactArgs(1),
	Short: "Get latest heartbeat for a service from Beacon",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			return err
		}
		return getStatus(server, args[0])
	},
}

func init() {
	heartbeatCmd.AddCommand(statusCmd)
}

func getStatus(server string, serviceId string) error {
	target := fmt.Sprintf("%s/status/%s", server, serviceId)
	resp, err := http.Get(target)
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
