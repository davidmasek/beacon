package monitor

import (
	"github.com/spf13/viper"
	"optimisticotter.me/heartbeat-monitor/storage"
)

type Monitor interface {
	Start(db storage.Storage, viper *viper.Viper) error
}
