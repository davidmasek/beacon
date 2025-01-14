package cmd

import (
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List known services",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		db, err := storage.InitDB(config.DbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		services, err := db.ListServices()
		if err != nil {
			return err
		}
		for _, service := range services {
			cmd.Println(service)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
