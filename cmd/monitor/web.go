// Run WebPinger. Check websites health and store to DB

package main

import (
	"fmt"

	"github.com/spf13/viper"
	"optimisticotter.me/heartbeat-monitor/monitor"
	"optimisticotter.me/heartbeat-monitor/storage"
)

func main() {
	viper := viper.New()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	websites := make(map[string]monitor.WebConfig)
	err = viper.UnmarshalKey("websites", &websites)
	if err != nil {
		panic(fmt.Errorf("fatal error unmarshaling config file: %w", err))
	}

	db, err := storage.InitDB()
	if err != nil {
		panic(err)
	}

	webPinger := monitor.WebPinger{}
	webPinger.Start(db, viper)
}
