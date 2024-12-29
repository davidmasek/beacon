package cmd

import (
	"log"
	"strings"

	"github.com/davidmasek/beacon/monitor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func loadConfig(cmd *cobra.Command) (*viper.Viper, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	if configFile != "" {
		return monitor.DefaultConfigFrom(configFile)
	}

	config, err := monitor.DefaultConfig()
	// TODO: quick fix to enable start when no config file found
	// default one should be created instead
	if err != nil && strings.Contains(err.Error(), `Config File "beacon.yaml" Not Found in`) {
		log.Println(err)
		cmd.Println("No config file found")
		err = nil
	}
	return config, err
}
