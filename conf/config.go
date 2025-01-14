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

	"github.com/caarlos0/env/v11"
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

func (s *Secret) IsSet() bool {
	return s.value != ""
}

type EmailConfig struct {
	SmtpServer   string `yaml:"smtp_server" env:"EMAIL_SMTP_SERVER"`
	SmtpPort     int    `yaml:"smtp_port" env:"EMAIL_SMTP_PORT"`
	SmtpUsername string `yaml:"smtp_username" env:"EMAIL_SMTP_USERNAME"`
	SmtpPassword Secret `yaml:"smtp_password" env:"EMAIL_SMTP_PASSWORD"`
	SendTo       string `yaml:"send_to" env:"EMAIL_SEND_TO"`
	Sender       string `yaml:"sender" env:"EMAIL_SENDER"`
	Prefix       string `yaml:"prefix" env:"EMAIL_PREFIX"`
}

func (s *Secret) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node got %s", node.Value)
	}
	s.value = node.Value
	return nil
}

type Config struct {
	// TODO: add examples to config + README
	Timezone time.Location `yaml:"timezone" env:"TIMEZONE"`
	// report after n-th hour in the day
	// e.g. 17 -> report after 5pm
	ReportAfter int `yaml:"report_after" env:"REPORT_AFTER"`

	EmailConf EmailConfig `yaml:"email"`

	services []ServiceConfig

	envPrefix string
	parents   []string
	settings  map[string]interface{}
	// manually set, should take precedence
	overrides map[string]interface{}
}

func (s Config) String() string {
	return fmt.Sprintf("BeaconConfig{\nEmail: %#v\nTimezone: %s\nReportAfter: %d\n}", s.EmailConf, s.Timezone.String(), s.ReportAfter)
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
	config.overrides[key] = value
}

func (config *Config) SetDefault(key string, value interface{}) {
	// todo: nested access
	if strings.Contains(key, ".") {
		panic("nested access with `.` not implemented")
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
		log.Printf("Key: %s\n", key)

		var err error
		switch key {
		case "services":
			config.services, err = parseServicesConfig(valueNode)
		case "email":
			// TODO: overwrite with env
			err = valueNode.Decode(&config.EmailConf)
		default:
			log.Printf("Unexpected key %v\n", key)
		}
		if err != nil {
			return err
		}
	}

	// TODO: use structured Config; currently quick "fix" to keep legacy code working
	err := value.Decode(config.settings)
	return err
}

// Parse config from YAML and override using ENV variables
func ConfigFromBytes(data []byte) (*Config, error) {
	config := NewConfig()
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	err = env.ParseWithOptions(config, env.Options{
		Prefix: ENV_VAR_PREFIX,
		OnSet: func(tag string, value interface{}, isDefault bool) {
			if value != nil {
				log.Printf("[env] Set %s\n", tag)
			}
		},
	})
	if err != nil {
		return nil, err
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
