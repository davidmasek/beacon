package cmd

import (
	"github.com/davidmasek/beacon/conf"
	"github.com/spf13/cobra"
)

func loadConfig(cmd *cobra.Command) (*conf.Config, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	if configFile != "" {
		return conf.DefaultConfigFrom(configFile)
	}

	config, err := conf.DefaultConfig()
	return config, err
}
