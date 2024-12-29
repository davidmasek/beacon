package monitor

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Setup default configuration.
// Tries to find config file automatically in home dir.
func DefaultConfig() (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigName("beacon.yaml")
	config.SetConfigType("yaml")
	// todo: is this portable?
	config.AddConfigPath("$HOME/")
	return setupConfig(config)
}

// Load config from "config.sample.yaml". Useful for testing.
func ExampleConfig() (*viper.Viper, error) {
	// test can be run with different working dir
	locations := []string{
		filepath.Join("config.sample.yaml"),
		filepath.Join("..", "config.sample.yaml"),
	}
	for _, loc := range locations {
		_, err := os.Stat(loc)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return DefaultConfigFrom(loc)
	}
	return nil, fmt.Errorf("config.sample.yaml file not found")
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
	log.Printf("read config from %q\n", config.ConfigFileUsed())

	// TODO: refactor
	// why does not Go have sets? :/
	keys := config.AllSettings()
	expectedKeys := []string{
		"services",
		"email",
	}
	for _, expected := range expectedKeys {
		if _, exists := keys[expected]; !exists {
			log.Printf("%q key not present in config", expected)
		}
		delete(keys, expected)
	}
	for key := range keys {
		log.Printf("unexpected %q key present in config", key)
	}

	return config, err
}
