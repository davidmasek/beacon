package cmd

import (
	"fmt"
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
		noMail, err := cmd.Flags().GetBool("no-mail")
		if err != nil {
			return err
		}

		if sendMail && noMail {
			return fmt.Errorf("cannot set both --send-mail and --no-mail")
		}

		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		if sendMail {
			// todo: figure out options for SendMail
			config.EmailConf.Enabled = "yes"
		}
		if noMail {
			config.EmailConf.Enabled = "no"
		}
		config.ReportName = reportName

		db, err := storage.InitDB(config.DbPath)
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
	reportCmd.Flags().Bool("no-mail", false, "Do not send email notifications")
	reportCmd.Flags().String("name", "report", "Report name")
}
