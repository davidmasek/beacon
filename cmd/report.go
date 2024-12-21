package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

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
		viper.SetConfigName("beacon.yaml")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/")
		err = viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("fatal error reading config file %q: %w", viper.ConfigFileUsed(), err)
		}
		log.Printf("Read config from %q", viper.ConfigFileUsed())
		viper.Set("send-mail", sendMail)
		viper.Set("report-name", reportName)

		return report(viper)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().Bool("send-mail", false, "Send email notifications")
	reportCmd.Flags().String("name", "report", "Report name")
}

func report(viper *viper.Viper) error {
	db, err := storage.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	reports, err := handlers.GenerateReport(db)
	if err != nil {
		return err
	}

	filename := viper.GetString("report-name")

	if !strings.HasSuffix(filename, ".html") {
		filename = fmt.Sprintf("%s.html", filename)
	}

	// proceed to send email even if writing to file fails
	// as it is better if at least one of the two succeeds
	err = handlers.WriteReportToFile(reports, filename)

	if viper.GetBool("send-mail") {
		var emailErr error
		server, emailErr := handlers.LoadServer(viper.Sub("email"))
		if emailErr != nil {
			emailErr := fmt.Errorf("failed to load SMTP server: %w", err)
			err = errors.Join(err, emailErr)
			return err
		}
		mailer := handlers.SMTPMailer{
			Server: server,
		}
		emailErr = mailer.Send(reports, viper.Sub("email"))
		err = errors.Join(err, emailErr)
	}

	return err
}
