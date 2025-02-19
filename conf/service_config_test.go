package conf

import (
	"fmt"
	"os"
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
  strawberry:
    url: "local.berry"
    token: "Dmh9Yr5rycRPHxb7nCDa"
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
	assert.Equal(t, true, services["beacon-periodic-checker"].Enabled)

	require.Contains(t, services, "strawberry")
	assert.Equal(t, true, services["strawberry"].Enabled)
	assert.Equal(t, "local.berry", services["strawberry"].Url)
	token := services["strawberry"].Token
	assert.Equal(t, "Dmh9Yr5rycRPHxb7nCDa", token.Get())
}

func TestReadTokenFromFile(t *testing.T) {
	configTemplate := `
services:
  strawberry:
    url: "local.berry"
    token: "Dmh9Yr5rycRPHxb7nCDa"
    token_file: %s
`
	file, err := os.CreateTemp("", "test_token")
	require.NoError(t, err)
	os.WriteFile(file.Name(), []byte("AaaDmh9Yr5rycRPHxb7nCDa\n"), 0644)

	configStr := fmt.Sprintf(configTemplate, file.Name())

	config, err := ConfigFromBytes([]byte(configStr))
	require.NoError(t, err)

	service := config.Services.Get("strawberry")
	require.NotNil(t, service)

	assert.Equal(t, "AaaDmh9Yr5rycRPHxb7nCDa", service.Token.Get())
	assert.Equal(t, "AaaDmh9Yr5rycRPHxb7nCDa", service.Token.FromFile)
	assert.Equal(t, "Dmh9Yr5rycRPHxb7nCDa", service.Token.Value)
}
