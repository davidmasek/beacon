package monitor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFrom(t *testing.T) {
	exampleConfigFile := filepath.Join("..", "config.sample.yaml")
	require.FileExists(t, exampleConfigFile)
	config, err := DefaultConfigFrom(exampleConfigFile)
	require.NoError(t, err)
	require.NotNil(t, config)
}
