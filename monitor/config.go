package monitor

import (
	"strings"

	"github.com/spf13/viper"
)

// Setup default configuration.
// Tries to find config file automatically (home dir, current dir).
func DefaultConfig() (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigName("beacon.yaml")
	config.SetConfigType("yaml")
	config.AddConfigPath(".")
	config.AddConfigPath("$HOME/")
	return setupConfig(config)
}

// Setup default configuration.
// Read config file from specified location.
func DefaultConfigFrom(configFile string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile(configFile)
	return setupConfig(config)
}

// Setup config to use ENV variables and read specified config file.
func setupConfig(config *viper.Viper) (*viper.Viper, error) {
	config.SetEnvPrefix("BEACON")
	// Bash doesn't allow dot in the environment variable name.
	// Viper requires dot for nested variables.
	// Use underscore and replace.
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// The combination of the prefix + string replacer
	// means that to overwrite config `email.smtp_port`, i.e.
	// ```yaml
	// email:
	//   smtp_port: 587
	// ```
	// you should use BEACON_EMAIL_SMTP_PORT key, e.g.
	// BEACON_EMAIL_SMTP_PORT=123
	config.AutomaticEnv()

	err := config.ReadInConfig()
	return config, err
}
