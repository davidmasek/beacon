package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TEST_CFG = []byte(`
services:
  beacon-github:
    url: "https://github.com/davidmasek/beacon"
    status:
      - 200
    content:
      - Beacon
  beacon-periodic-checker:
  example-basic-web:
    url: "https://httpbin.org/get"
  example-temp-disable:
    url: "will-not-be-used-because-disabled"
    enabled: false
`)

func TestExampleConfigServices(t *testing.T) {
	config, err := ConfigFromBytes(TEST_CFG)
	require.NoError(t, err)

	services := map[string]ServiceConfig{}
	for _, cfg := range config.AllServices() {
		services[cfg.Id] = cfg
	}

	assert.Equal(t, true, services["beacon-github"].Enabled)
	assert.Equal(t, []int{200}, services["beacon-github"].HttpStatus)
	assert.Equal(t, 24*time.Hour, services["beacon-github"].Timeout)

	assert.Equal(t, "https://httpbin.org/get", services["example-basic-web"].Url)
	// test default value (200 OK) gets assigned
	assert.Equal(t, []int{200}, services["example-basic-web"].HttpStatus)
	assert.Nil(t, services["example-basic-web"].BodyContent)

	assert.Equal(t, false, services["example-temp-disable"].Enabled)

	// Service without config should still be included.
	require.Contains(t, services, "beacon-periodic-checker")
	assert.Equal(t, services["beacon-periodic-checker"].Enabled, true)
}
