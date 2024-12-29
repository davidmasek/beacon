package cmd

import (
	"time"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Args:  cobra.ExactArgs(0),
	Short: "Generate and optionally send report about current service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		sendMail, err := cmd.Flags().GetBool("send-mail")
		if err != nil {
			return err
		}
		reportName, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		config.Set("send-mail", sendMail)
		config.Set("report-name", reportName)

		db, err := storage.InitDB(config.GetString("DB"))
		if err != nil {
			return err
		}
		defer db.Close()
		return handlers.DoReportTask(db, config, time.Now())
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().Bool("send-mail", false, "Send email notifications")
	reportCmd.Flags().String("name", "report", "Report name")
}
