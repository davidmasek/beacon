package conf

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
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
	SmtpServer   string `yaml:"smtp_server" env:"SMTP_SERVER"`
	SmtpPort     int    `yaml:"smtp_port" env:"SMTP_PORT"`
	SmtpUsername string `yaml:"smtp_username" env:"SMTP_USERNAME"`
    SmtpPassword Secret `yaml:"smtp_password" env:"SMTP_PASSWORD"`
	SendTo       string `yaml:"send_to" env:"SEND_TO"`
	Sender       string `yaml:"sender" env:"SENDER"`
	Prefix       string `yaml:"prefix" env:"PREFIX"`
	// not bool to allow more flexible usage
	Enabled string `yaml:"enabled" env:"ENABLED"`
}

func (emailConf *EmailConfig) IsEnabled() bool {
	// explicitly enabled
	if emailConf.Enabled == "yes" || emailConf.Enabled == "true" {
		return true
	}
	// explicitly disabled
	if emailConf.Enabled == "no" || emailConf.Enabled == "false" {
		return false
	}
	return emailConf.SmtpPassword.IsSet()
}

func (s *Secret) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node got %s", node.Value)
	}
	s.value = node.Value
	return nil
}

// TzLocation wraps a *time.Location so we can provide custom YAML unmarshalling.
type TzLocation struct {
	Location *time.Location
}

type Config struct {
	// TODO: add examples to config + README
	Timezone TzLocation `yaml:"timezone" env:"TIMEZONE"`
	// report after n-th hour in the day
	// e.g. 17 -> report after 5pm
	ReportAfter     int           `yaml:"report_time" env:"REPORT_TIME"`
	DbPath          string        `yaml:"db_path" env:"DB"`
	ReportName      string        `yaml:"report_name" env:"REPORT_NAME"`
	Port            int           `yaml:"port" env:"PORT"`
	SchedulerPeriod time.Duration `yaml:"scheduler_period" env:"SCHEDULER_PERIOD"`

    EmailConf EmailConfig `yaml:"email" envPrefix:"EMAIL_"`

	Services ServicesList

	envPrefix string
}

func (s Config) AllServices() []ServiceConfig {
	return s.Services.Services
}

func (s Config) String() string {
	confStr, err := yaml.Marshal(s)
	if err != nil {
		return err.Error()
	}
	nServices := -1
	if s.Services.Services != nil {
		nServices = len(s.Services.Services)
	}
	return fmt.Sprintf("BeaconConfig[%d services]\n%s\n", nServices, confStr)
	// return fmt.Sprintf("BeaconConfig{\nEmail: %#v\nTimezone: %s\nReportAfter: %d\n}", s.EmailConf, s.Timezone.String(), s.ReportAfter)
}

func (s Config) GoString() string {
	return "Config{*****}"
}

// TODO: keep this as Config strict and if needed to write than marshal?
//
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
		Timezone: TzLocation{time.Local},
		// todo[defaults]: better defaults approach
		ReportAfter:     17,
		Port:            8088,
		SchedulerPeriod: 15 * time.Minute,
		envPrefix:       ENV_VAR_PREFIX,
	}
	config.Services.Services = []ServiceConfig{}
	return config
}

func (tz *TzLocation) UnmarshalYAML(value *yaml.Node) error {
	var tzName string
	if err := value.Decode(&tzName); err != nil {
		return err
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return fmt.Errorf("failed to load location %q: %w", tzName, err)
	}
	tz.Location = loc
	return nil
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
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(Secret{}): func(v string) (interface{}, error) {
				return Secret{v}, nil
			},
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
