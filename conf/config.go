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

type Config struct {
	envPrefix string
	settings  map[string]interface{}
	// manually set, should take precedence
	overrides map[string]interface{}
}

func (config *Config) AllSettings() map[string]interface{} {
	return config.settings
}

func (config *Config) keyToEnvVar(key string) string {
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToUpper(key)
	return config.envPrefix + key
}

func (config *Config) get(key string) interface{} {
	// TODO: dot notation for nested access not implemented
	val, ok := config.overrides[key]
	if ok {
		return val
	}
	// overwrite with ENV var if available
	envVal, isSet := os.LookupEnv(config.keyToEnvVar(key))
	if isSet {
		return envVal
	}
	val, ok = config.settings[key]
	// TODO: do we want default or nil?
	if !ok {
		return nil
	}
	return val
}

func (config *Config) GetString(key string) string {
	val := config.get(key)
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
	val := config.settings[key]
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
	// TODO: dot notation for nested access not implemented
	config.overrides[key] = value
}

func (config *Config) SetDefault(key string, value interface{}) {
	if !config.IsSet(key) {
		config.settings[key] = value
	}
}

func (config *Config) IsSet(key string) bool {
	_, ok := config.overrides[key]
	if ok {
		return true
	}
	_, isSet := os.LookupEnv(config.keyToEnvVar(key))
	if isSet {
		return true
	}
	_, ok = config.settings[key]
	return ok
}

func subSettings(parent map[string]interface{}, key string) map[string]interface{} {
	val, ok := parent[key]
	if !ok {
		return map[string]interface{}{}
	}
	child, ok := val.(map[string]interface{})
	// if config[key] is not a dict, it is a scalar
	// to keep the same structure we convert it to map
	// with the scalar as key and no value
	// todo: maybe better design could be pondered here
	if !ok {
		child = make(map[string]interface{})
		child[val.(string)] = struct{}{}
	}
	return child
}

func (config *Config) Sub(key string) *Config {
	// TODO: doing a quick implementation here
	// either needs more testing or remove this method
	sub := &Config{}
	sub.envPrefix = config.envPrefix + strings.ToUpper(key) + "_"
	sub.overrides = subSettings(config.overrides, key)
	sub.settings = subSettings(config.settings, key)

	return sub
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
	return setupConfig(configFile)
}

// Empty config
func NewConfig() *Config {
	config := &Config{
		envPrefix: ENV_VAR_PREFIX,
		settings:  make(map[string]interface{}),
		overrides: make(map[string]interface{}),
	}
	return config
}

// Minimal working config
func BaseConfig() *Config {
	config := NewConfig()
	config.settings["services"] = []string{}
	config.settings["email"] = map[string]string{}
	return config
}

// Setup config to use ENV variables and read specified config file.
func setupConfig(configFile string) (*Config, error) {
	log.Printf("reading config from %q\n", configFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := NewConfig()
	err = yaml.Unmarshal(data, config.settings)
	if err != nil {
		return nil, err
	}
	log.Println(">>>>", config, "<<<<")
	return config, err
}
