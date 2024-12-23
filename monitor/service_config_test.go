package monitor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadExampleConfig(t *testing.T) {
	configFile := filepath.Join("..", "config.sample.yaml")
	config, err := DefaultConfigFrom(configFile)
	require.NoError(t, err)

	services, err := ParseServiceConfig(config.Sub("services"))
	require.NoError(t, err)

	for id, service := range services {
		t.Log(id, service)
	}
}
