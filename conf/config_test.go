package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoadConfigFrom(t *testing.T) {
	exampleConfigFile := filepath.Join("..", "config.sample.yaml")
	require.FileExists(t, exampleConfigFile)
	config, err := DefaultConfigFrom(exampleConfigFile)
	require.NoError(t, err)
	require.NotNil(t, config)

	require.True(t, config.IsSet("services"), config)
	require.True(t, config.IsSet("email"), config)
}

func TestEnvVariablesOverwrite(t *testing.T) {
	err := os.Setenv("BEACON_EMAIL_PREFIX", "my-new-prefix")
	require.NoError(t, err)
	err = os.Setenv("BEACON_EMAIL_SMTP_PORT", "123")
	require.NoError(t, err)

	exampleConfigFile := filepath.Join("..", "config.sample.yaml")
	require.FileExists(t, exampleConfigFile)
	config, err := DefaultConfigFrom(exampleConfigFile)
	require.NoError(t, err)
	require.NotNil(t, config)

	require.True(t, config.IsSet("email"), config)
	emailConfig := config.Sub("email")
	t.Log(emailConfig.AllSettings())
	assert.Equal(t, "my-new-prefix", emailConfig.GetString("prefix"), emailConfig)
	assert.Equal(t, 123, emailConfig.GetInt("smtp_port"), emailConfig)
}

func TestConfigGet(t *testing.T) {
	config := NewConfig()
	data := `
bedroom: bed
kitchen:
  fruit: apple
  vegetable: cucumber
  table:
`
	err := yaml.Unmarshal([]byte(data), config.settings)
	require.NoError(t, err)
	require.NotNil(t, config)
	t.Log(config)

	require.True(t, config.IsSet("bedroom"))
	require.True(t, config.IsSet("kitchen"))
	require.Equal(t, "bed", config.GetString("bedroom"))

	settings := config.AllSettings()
	require.Contains(t, settings, "bedroom")
	require.Contains(t, settings, "kitchen")

	kitchen := config.get("kitchen")
	require.NotNil(t, kitchen)
	t.Log(kitchen)

	kitchenConfig := config.Sub("kitchen")
	require.NotNil(t, kitchenConfig)
	t.Log(kitchenConfig)

	require.True(t, kitchenConfig.IsSet("fruit"))
	require.True(t, kitchenConfig.IsSet("vegetable"))
	require.Equal(t, "apple", kitchenConfig.GetString("fruit"))
	// todo: Config rethink... the following will return false by design
	// but the key exists!
	// require.True(t, kitchenConfig.IsSet("table"))
	// but for our use case it is more important that the following works:
	settings = kitchenConfig.AllSettings()
	require.Contains(t, settings, "fruit")
	require.Contains(t, settings, "vegetable")
	require.Contains(t, settings, "table")
}

func TestConfigSet(t *testing.T) {
	config := NewConfig()
	config.Set("foo", true)
	foo := config.GetBool("foo")
	require.True(t, foo)
}

func TestSecretPrint(t *testing.T) {
	secret := Secret{"Greg"}
	assert.Equal(t, secret.Get(), "Greg")
	assert.NotContains(t, fmt.Sprint(secret), "Greg")
	assert.NotContains(t, fmt.Sprint(&secret), "Greg")
	assert.NotContains(t, fmt.Sprint([]Secret{secret}), "Greg")
	assert.NotContains(t, fmt.Sprint([]*Secret{&secret}), "Greg")
	assert.NotContains(t, secret.String(), "Greg")
	assert.NotContains(t, fmt.Sprintf("%v", secret), "Greg")
	assert.NotContains(t, fmt.Sprintf("%+v", secret), "Greg")
	assert.NotContains(t, fmt.Sprintf("%#v", secret), "Greg")

	config := NewConfig()
	config.settings["password"] = secret
	assert.NotContains(t, fmt.Sprint(config), "Greg")
	assert.NotContains(t, fmt.Sprintf("%v", config), "Greg")
	assert.NotContains(t, fmt.Sprintf("%+v", config), "Greg")
	assert.NotContains(t, fmt.Sprintf("%#v", config), "Greg")
}
