package cmd

import (
	"fmt"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use: "debug",
	// Args:  cobra.ExactArgs(0),
	Short: "Development tools, use with caution",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.Get()
		logger.Warn("Debug CLI - Use with caution")
		deleteTasks, err := cmd.Flags().GetBool("delete-tasks")
		if err != nil {
			return err
		}
		listSchema, err := cmd.Flags().GetBool("schema")
		if err != nil {
			return err
		}
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		db, err := storage.InitDB(config.DbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()
		if listSchema {
			cmd.Println("DB Schema:")
			schemas, err := db.ListSchemaVersions()
			if err != nil {
				return fmt.Errorf("failed to list schemas: %w", err)
			}
			for _, schema := range schemas {
				cmd.Printf("- %#v\n", schema)
			}
		}
		if deleteTasks {
			err = db.DropTasks()
			if err != nil {
				return fmt.Errorf("failed to drop tasks: %w", err)
			}
		}

		cmd.Println("Done")
		return nil
	},
}

func init() {
	debugCmd.Flags().Bool("delete-tasks", false, "")
	debugCmd.Flags().Bool("schema", false, "")

	rootCmd.AddCommand(debugCmd)
}
