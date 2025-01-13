package conf

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// prefix for environment variables
const ENV_VAR_PREFIX = "BEACON_"

type Secret struct {
	value string
}

func (s Secret) String() string {
	return "*****"
}

func (s Secret) GoString() string {
	return "Secret{*****}"
}

func (s *Secret) Get() string {
	return s.value
}

type Config struct {
	// TODO: expose in config file
	Timezone time.Location
	// TODO: expose in config file
	// report after n-th hour in the day
	// e.g. 17 -> report after 5pm
	ReportAfter int

	services []ServiceConfig

	envPrefix string
	parents   []string
	settings  map[string]interface{}
	// manually set, should take precedence
	overrides map[string]interface{}
}

func (s Config) String() string {
	// todo: should actually print something useful,
	// but at least this does not leak secrets
	return fmt.Sprintf("BeaconConfig%s", s.parents)
}

func (s Config) GoString() string {
	return "Config{*****}"
}

func (s Config) Services() []ServiceConfig {
	return s.services
}

func (config *Config) AllSettings() map[string]interface{} {
	// todo: does not use overrides
	// - this is OK for the current use-case but terrible
	// if used in other ways
	// - Config needs refactor anyway, so not dealing with this now
	// - probably the whole method should be removed in future
	settings := config.settings
	for _, parent := range config.parents {
		if settings == nil {
			return nil
		}
		settingsSub, ok := settings[parent].(map[string]interface{})
		if ok {
			settings = settingsSub
		} else {
			return nil
		}
	}
	return settings
}

func (config *Config) keyToEnvVar(key string) string {
	// todo: nested access
	if strings.Contains(key, ".") {
		panic("nested access with `.` not implemented")
	}
	// key = strings.ReplaceAll(key, ".", "_")
	key = config.envPrefix + strings.Join(config.parents, "_") + "_" + key
	key = strings.ToUpper(key)
	return key
}

func (config *Config) get(key string) interface{} {
	if config == nil {
		return nil
	}
	// todo: nested access
	if strings.Contains(key, ".") {
		panic("nested access with `.` not implemented")
	}
	overrides := config.overrides
	for _, parent := range config.parents {
		if overrides == nil {
			break
		}
		overridesSub, ok := overrides[parent].(map[string]interface{})
		if ok {
			overrides = overridesSub
		}
	}
	val, ok := overrides[key]
	if ok {
		return val
	}
	// overwrite with ENV var if available
	envVal, isSet := os.LookupEnv(config.keyToEnvVar(key))
	if isSet {
		return envVal
	}
	settings := config.settings
	for _, parent := range config.parents {
		if settings == nil {
			return nil
		}
		settingsSub, ok := settings[parent].(map[string]interface{})
		if ok {
			settings = settingsSub
		} else {
			return nil
		}
	}
	val, ok = settings[key]

	if !ok {
		return nil
	}
	return val
}

func (config *Config) GetString(key string) string {
	val := config.get(key)
	if val == nil {
		return ""
	}
	strVal, ok := val.(string)
	if ok {
		return strVal
	}
	return fmt.Sprint(val)
}

func (config *Config) GetInt(key string) int {
	val := config.get(key)
	intVal, ok := val.(int)
	if ok {
		return intVal
	}
	strVal, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("For key %q - cannot convert %q to int", key, val))
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		panic(fmt.Sprintf("For key %q - cannot convert %q to int", key, val))
	}
	return intVal
}

var boolyStrings = map[string]bool{
	"true":  true,
	"1":     true,
	"TRUE":  true,
	"false": false,
	"0":     false,
	"FALSE": false,
}

func (config *Config) GetBool(key string) bool {
	val := config.get(key)
	boolVal, isBool := val.(bool)
	if isBool {
		return boolVal
	}
	strVal, isString := val.(string)
	if !isString {
		panic(fmt.Sprintf("Cannot parse key %q with value %q as bool", key, config.settings[key]))
	}
	boolVal, isExpectedFormat := boolyStrings[strVal]
	if !isExpectedFormat {
		panic(fmt.Sprintf("Cannot parse key %q with value %q as bool", key, config.settings[key]))
	}
	return boolVal
}

func (config *Config) GetDuration(key string) time.Duration {
	value := config.get(key)
	durationValue, isDuration := value.(time.Duration)
	if isDuration {
		return durationValue
	}
	parsedValue, err := time.ParseDuration(value.(string))
	if err != nil {
		panic(fmt.Sprintf("Cannot parse %q as time.Duration", value))
	}
	return parsedValue
}

func (config *Config) Set(key string, value interface{}) {
	// todo: nested access
	if strings.Contains(key, ".") {
		panic("nested access with `.` not implemented")
	}
	if len(config.parents) > 0 {
		// todo: sub configs are read only for now
		// not sure what to do with them atm
		panic("Cannot set values for .Sub configs")
	}
	config.overrides[key] = value
}

func (config *Config) SetDefault(key string, value interface{}) {
	// todo: nested access
	if strings.Contains(key, ".") {
		panic("nested access with `.` not implemented")
	}
	if len(config.parents) > 0 {
		// todo: sub configs are read only for now
		// not sure what to do with them atm
		panic("Cannot set values for .Sub configs")
	}
	val := config.get(key)
	if val == nil {
		config.settings[key] = value
	}
}

func (config *Config) IsSet(key string) bool {
	// here if the key exists but has nil value we return false
	// now this is kinda stupid but it kinda makes sense for our use-cases
	// I don't have a solution that would be simple to do and work well atm
	// todo: probably want to rethink the whole Config anyway
	val := config.get(key)
	return val != nil
}

func (config *Config) Sub(key string) *Config {
	// todo: kinda weird implementation, not sure how I want to use this yet
	if !config.IsSet(key) {
		return nil
	}
	return &Config{
		envPrefix: config.envPrefix,
		parents:   append(config.parents, key),
		settings:  config.settings,
		overrides: config.overrides,
	}
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
func DefaultConfig() (*Config, error) {
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
func ExampleConfig() (*Config, error) {
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
func DefaultConfigFrom(configFile string) (*Config, error) {
	err := ensureConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	return configFromFile(configFile)
}

// Empty config
func NewConfig() *Config {
	config := &Config{
		Timezone: *time.Local,
		// todo: better defaults approach
		ReportAfter: 17,
		envPrefix:   ENV_VAR_PREFIX,
		services:    []ServiceConfig{},
		parents:     []string{},
		settings:    make(map[string]interface{}),
		overrides:   make(map[string]interface{}),
	}
	return config
}

func (config *Config) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", value.Kind)
	}

	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]

		// Access the key and value
		key := keyNode.Value
		fmt.Printf("Key: %s\n", key)

		var err error
		switch key {
		case "services":
			config.services, err = parseServicesConfig(valueNode)
			if err != nil {
				return err
			}
		default:
			fmt.Printf("Unexpected key %v\n", key)
		}
	}

	// TODO: use structured Config; currently quick "fix" to keep legacy code working
	err := value.Decode(config.settings)
	return err
}

func ConfigFromBytes(data []byte) (*Config, error) {
	config := NewConfig()
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	// todo: improve; quick hack to hide SMTP password in logging
	if config.IsSet("email") {
		emailConfig := config.Sub("email")
		if emailConfig.IsSet("smtp_password") {
			config.settings["email"].(map[string]any)["smtp_password"] = Secret{
				value: emailConfig.GetString("smtp_password"),
			}
		}
	}
	log.Println(">>>>", config, "<<<<")
	return config, err
}

func configFromFile(configFile string) (*Config, error) {
	log.Printf("reading config from %q\n", configFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return ConfigFromBytes(data)
}
