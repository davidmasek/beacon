package cmd

import (
	"fmt"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		// TODO: look into how viper/cobra should be used together
		viper := viper.New()
		viper.AddConfigPath(".")
		err = viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("fatal error reading config file: %w", err)
		}
		viper.Set("send-mail", sendMail)
		viper.Set("report-name", reportName)

		return report(viper)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().Bool("send-mail", false, "Send email notifications")
	reportCmd.Flags().String("name", "dev-report", "Report name")
}

func report(viper *viper.Viper) error {
	var mailer handlers.Mailer

	if viper.GetBool("send-mail") {
		server, err := handlers.LoadServer(viper.Sub("email"))
		if err != nil {
			return fmt.Errorf("failed to load SMTP server: %w", err)
		}
		mailer = handlers.SMTPMailer{
			Server: server,
		}
	} else {
		mailer = handlers.FakeMailer{}
	}

	db, err := storage.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	reports, err := handlers.GenerateReport(db)
	if err != nil {
		return err
	}

	return mailer.Send(reports, viper.Sub("email"))
}
