package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/davidmasek/beacon/conf"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Args:  cobra.ExactArgs(0),
	Short: "Print version/build information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("cannot read build info")
		}
		cmd.Println("Beacon version:", info.Main.Version)
		cmd.Println("Manually-specified version:", conf.GitRef, conf.GitSha)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
