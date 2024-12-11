package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status service-id",
	Args:  cobra.ExactArgs(1),
	Short: "Get service status from Beacon",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			return err
		}
		response, err := getStatus(server, args[0])
		fmt.Println(response)
		return err
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func getStatus(server string, serviceId string) (string, error) {
	target := fmt.Sprintf("%s/status/%s", server, serviceId)
	resp, err := http.Get(target)
	if err != nil {
		log.Println("Cannot Get", target)
		return "", err
	}
	if resp == nil {

		return "", errors.New("No response")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Cannot Read response", body, err)
		return "", err
	}
	return string(body), nil
}
