package conf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Load a few example timezone locations
func TestLoadTimezone(t *testing.T) {
	_, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	_, err = time.LoadLocation("Australia/Sydney")
	require.NoError(t, err)
	_, err = time.LoadLocation("America/Chicago")
	require.NoError(t, err)
}

func TestEnvVariablesOverwrite(t *testing.T) {
	err := os.Setenv("BEACON_EMAIL_PREFIX", "my-new-prefix")
	require.NoError(t, err)
	err = os.Setenv("BEACON_EMAIL_SMTP_PORT", "123")
	require.NoError(t, err)

	data := []byte(`
email:
  smtp_port: 4554
  prefix: "[beacon-example]"
`)
	config, err := ConfigFromBytes(data)
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

testing: true

timezone: "Europe/Prague"
report_time: 17
scheduler_period: 33m
`)
	err := yaml.Unmarshal(data, config)
	require.NoError(t, err)
	require.Equal(t, "david", config.EmailConf.SmtpUsername)

	expectedTimezone, err := time.LoadLocation("Europe/Prague")
	require.NoError(t, err)
	// todo[defaults]: would prefer not having constants here
	// but need some testing for parsing the more complex types
	// Should refactor this once we have better defaults
	assert.Equal(t, expectedTimezone, config.Timezone.Location,
		fmt.Sprintf("%s x %s", expectedTimezone.String(), config.Timezone.Location.String()))
	assert.Equal(t, 17, config.ReportAfter)
	assert.Equal(t, 33*time.Minute, config.SchedulerPeriod)
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
	names := []string{}
	for _, service := range config.AllServices() {
		names = append(names, service.Id)
	}
	require.Equal(t, expectedNames, names)
}

func TestSecretPrint(t *testing.T) {
	secret := Secret{"Greg", ""}
	assert.Equal(t, secret.Get(), "Greg")
	assert.NotContains(t, fmt.Sprint(secret), "Greg")
	assert.NotContains(t, fmt.Sprint(&secret), "Greg")
	assert.NotContains(t, fmt.Sprint([]Secret{secret}), "Greg")
	assert.NotContains(t, fmt.Sprint([]*Secret{&secret}), "Greg")
	assert.NotContains(t, secret.String(), "Greg")
	assert.NotContains(t, fmt.Sprintf("%v", secret), "Greg")
	assert.NotContains(t, fmt.Sprintf("%+v", secret), "Greg")
	assert.NotContains(t, fmt.Sprintf("%#v", secret), "Greg")

	config, err := ConfigFromBytes([]byte(`email:
  smtp_password: Greg
`))
	require.NoError(t, err)
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
		Secret{"h4xor", ""},
		"you@example.fake",
		"noreply@example.fake",
		"[test]",
		"",
		"",
	})
	require.Equal(t, "h4xor", emailConfig.SmtpPassword.Get())
}

func TestSecretFromEnv(t *testing.T) {
	err := os.Setenv("BEACON_EMAIL_SMTP_PASSWORD", "secr4t")
	require.NoError(t, err)
	conf, err := ConfigFromBytes([]byte(""))
	require.NoError(t, err)
	require.Equal(t, "secr4t", conf.EmailConf.SmtpPassword.Get())

}

func TestSecretFromFile(t *testing.T) {
	// Create the file
	f, err := os.Create("secret-test.txt")
	require.NoError(t, err)
	_, err = f.WriteString("foo\n")
	require.NoError(t, err)

	// Set both, file should have prio
	err = os.Setenv("BEACON_EMAIL_SMTP_PASSWORD", "bar")
	require.NoError(t, err)
	err = os.Setenv("BEACON_EMAIL_SMTP_PASSWORD_FILE", "secret-test.txt")
	require.NoError(t, err)

	conf, err := ConfigFromBytes([]byte(""))
	require.NoError(t, err)
	require.Equal(t, "foo", conf.EmailConf.SmtpPassword.Get())

	// Cleanup
	err = os.Unsetenv("BEACON_EMAIL_SMTP_PASSWORD_FILE")
	require.NoError(t, err)
	err = os.Remove("secret-test.txt")
	require.NoError(t, err)
}
