package cmd

import (
	"log"
	"strings"

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
	// TODO: quick fix to enable start when no config file found
	// default one should be created instead
	if err != nil && strings.Contains(err.Error(), `Config File "beacon.yaml" Not Found in`) {
		log.Println(err)
		cmd.Println("No config file found")
		err = nil
	}
	return config, err
}
