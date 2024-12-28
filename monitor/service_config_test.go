package monitor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExampleConfig(t *testing.T) {
	configFile := filepath.Join("..", "config.sample.yaml")
	config, err := DefaultConfigFrom(configFile)
	require.NoError(t, err)

	services, err := ParseServicesConfig(config.Sub("services"))
	require.NoError(t, err)

	// TODO: these are in the wrong order
	// should be "expected", "actual"
	assert.Equal(t, services["beacon-github"].Enabled, true)
	assert.Equal(t, services["beacon-github"].HttpStatus, []int{200})
	assert.Equal(t, services["beacon-github"].Timeout, 24*time.Hour)

	assert.Equal(t, services["example-with-extras"].Timeout, 48*time.Hour)

	assert.Equal(t, services["example-basic-web"].Url, "https://httpbin.org/get")
	// test default value (200 OK) gets assigned
	assert.Equal(t, []int{200}, services["example-basic-web"].HttpStatus)
	assert.Nil(t, services["example-basic-web"].BodyContent)

	assert.Equal(t, services["example-temp-disable"].Enabled, false)
}
