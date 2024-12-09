package monitor

import (
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

type Monitor interface {
	Start(db storage.Storage, viper *viper.Viper) error
}
