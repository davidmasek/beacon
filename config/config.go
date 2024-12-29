package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	settings map[string]interface{}
}

//go:embed config.default.yaml
var DEFAULT_CONFIG []byte

func ensureConfigFile(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.WriteFile(path, DEFAULT_CONFIG, 0644)
		return err
	}
	return err
}

// Load config file from home dir (such as `~/beacon.yaml`).
//
// Create config file if not found.
// Setup config to use env variables.
func DefaultConfig() (*viper.Viper, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(homedir, "beacon.yaml")
	err = ensureConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	return DefaultConfigFrom(configFile)
}

// Load config file from `config.sample.yaml`. Useful for testing.
//
// Fail if example config file not found.
// Setup config to use env variables.
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
		if err != nil {
			return nil, err
		}
		return DefaultConfigFrom(loc)
	}
	return nil, fmt.Errorf("config.sample.yaml file not found")
}

// Load config file from the specified path.
//
// Create config file if not found.
// Setup config to use env variables.
func DefaultConfigFrom(configFile string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile(configFile)
	err := ensureConfigFile(configFile)
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(config.ConfigFileUsed())
	if err != nil {
		return nil, err
	}
	config2 := &Config{settings: make(map[string]interface{})}
	err = yaml.Unmarshal(data, config2.settings)
	log.Println(">>>>", config2, "<<<<")
	fmt.Printf("Config: %+v\n", config2.settings)

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
