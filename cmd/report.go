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
		configFile, err := cmd.Flags().GetString("config-file")
		if err != nil {
			return err
		}

		// TODO: look into how viper/cobra should be used together
		config := viper.New()
		if configFile != "" {
			config.SetConfigFile(configFile)
		} else {
			config.SetConfigName("beacon.yaml")
			config.SetConfigType("yaml")
			config.AddConfigPath(".")
			config.AddConfigPath("$HOME/")
		}

		config.SetEnvPrefix("BEACON")
		// Bash doesn't allow dot in the environment variable name.
		// Viper requires dot for nested variables.
		// Use underscore and replace.
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()

		err = config.ReadInConfig()
		if err != nil {
			return fmt.Errorf("error reading config file %q: %w", config.ConfigFileUsed(), err)
		}
		log.Printf("Read config from %q", config.ConfigFileUsed())
		config.Set("send-mail", sendMail)
		config.Set("report-name", reportName)

		return report(config)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().Bool("send-mail", false, "Send email notifications")
	reportCmd.Flags().String("name", "report", "Report name")
}

func report(config *viper.Viper) error {
	db, err := storage.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	reports, err := handlers.GenerateReport(db)
	if err != nil {
		return err
	}

	filename := config.GetString("report-name")

	if !strings.HasSuffix(filename, ".html") {
		filename = fmt.Sprintf("%s.html", filename)
	}

	// proceed to send email even if writing to file fails
	// as it is better if at least one of the two succeeds
	err = handlers.WriteReportToFile(reports, filename)

	if config.GetBool("send-mail") {
		var emailErr error
		server, emailErr := handlers.LoadServer(config.Sub("email"))
		if emailErr != nil {
			emailErr := fmt.Errorf("failed to load SMTP server: %w", err)
			err = errors.Join(err, emailErr)
			return err
		}
		mailer := handlers.SMTPMailer{
			Server: server,
		}
		emailErr = mailer.Send(reports, config.Sub("email"))
		err = errors.Join(err, emailErr)
	}

	return err
}
