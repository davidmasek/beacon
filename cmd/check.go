package cmd

import (
	"strings"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check service-id url",
	Args:  cobra.ExactArgs(2),
	Short: "Check website status",
	RunE: func(cmd *cobra.Command, args []string) error {
		expectStatus, err := cmd.Flags().GetIntSlice("status")
		if err != nil {
			return err
		}
		expectContent, err := cmd.Flags().GetStringArray("content")
		if err != nil {
			return err
		}

		url := args[1]
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}

		checkConfig := monitor.WebConfig{
			Url:         url,
			HttpStatus:  expectStatus,
			BodyContent: expectContent,
		}

		serviceId := args[0]

		return checkWeb(serviceId, checkConfig)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().IntSlice("status", []int{200},
		"status code to accept (you can provide multiple with multiple flags or as comma separated list)")
	checkCmd.Flags().StringArray("content", []string{},
		"required string in result to mark check as success (you can provide multiple with multiple flags)")
}

func checkWeb(serviceId string, checkConfig monitor.WebConfig) error {
	db, err := storage.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	websites := make(map[string]monitor.WebConfig)
	websites[serviceId] = checkConfig

	return monitor.CheckWebsites(db, websites)
}
