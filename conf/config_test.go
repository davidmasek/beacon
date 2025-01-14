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

	emailConfig := config.EmailConf
	t.Log(emailConfig)
	t.Log("^^^^^^^^^^^^^^^^^^^^^^")
	assert.Equal(t, "my-new-prefix", emailConfig.Prefix)
	assert.Equal(t, 123, emailConfig.SmtpPort)
}

func TestUnmarshal(t *testing.T) {
	config := NewConfig()
	data := []byte(`
services:
  foo:
  bar:
    enabled: true

email:
  smtp_username: david

testing: true`)
	err := yaml.Unmarshal(data, config)
	require.NoError(t, err)
	require.Equal(t, "david", config.EmailConf.SmtpUsername)
}

func TestParsingKeepsOrder(t *testing.T) {
	expectedNames := []string{"foo", "bar", "third", "last"}
	data := []byte(`
services:
  foo:
  bar:
  third:
    url: ""
  last:`)
	config, err := ConfigFromBytes(data)
	require.NoError(t, err)
	services := config.Services()
	names := []string{}
	for _, service := range services {
		names = append(names, service.Id)
	}
	require.Equal(t, expectedNames, names)
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

func TestParseEmailConfig(t *testing.T) {
	src := `
smtp_server: mail.smtp2go.com
smtp_port: 587
smtp_username: beacon
smtp_password: h4xor
send_to: you@example.fake
sender: noreply@example.fake
prefix: "[test]"
`
	emailConfig := EmailConfig{}
	err := yaml.Unmarshal([]byte(src), &emailConfig)
	require.NoError(t, err)
	t.Logf("%#v\n", emailConfig)
	require.NotContains(t, fmt.Sprintf("%#v", emailConfig), "h4xor",
		"Password (secret) value should not be logged")
	require.Equal(t, emailConfig, EmailConfig{
		"mail.smtp2go.com",
		587,
		"beacon",
		Secret{"h4xor"},
		"you@example.fake",
		"noreply@example.fake",
		"[test]",
	})
	require.Equal(t, "h4xor", emailConfig.SmtpPassword.Get())
}
