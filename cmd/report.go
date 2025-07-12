package cmd

import (
	"time"

	"github.com/davidmasek/beacon/reporting"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Args:  cobra.ExactArgs(0),
	Short: "Generate and send report about current service status",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		db, err := storage.InitDB(config.DbPath)
		if err != nil {
			return err
		}
		defer func() {
			closeErr := db.Close()
			if err != nil {
				err = closeErr
			}
		}()
		reports, err := reporting.GenerateReport(db, config)
		if err != nil {
			return err
		}
		return reporting.SaveSendReport(reports, db, config, time.Now())
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
