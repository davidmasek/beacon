package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "beacon",
	Short: "Beacon: Simple heartbeat and website monitor.",
	Long: `Beacon: Track health of your projects. 
	
	This CLI enables you to starting individual Beacon components,
	manage your projects, and communicate with Beacon.

	https://github.com/davidmasek/beacon
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("server", "http://localhost:8088", "Address of the target heartbeat server")
	rootCmd.PersistentFlags().String("config", "", "Path to config file. If not specified, looks for beacon.yaml inside current and home directory.")
}
