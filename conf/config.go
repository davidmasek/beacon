package conf

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/davidmasek/beacon/logging"
	"gopkg.in/yaml.v3"
)

// prefix for environment variables
const ENV_VAR_PREFIX = "BEACON_"

type Secret struct {
	Value    string `env:""`
	FromFile string `env:"_FILE,file"`
}

func (s Secret) String() string {
	return "*****"
}

func (s Secret) GoString() string {
	return "Secret{*****}"
}

func (s *Secret) Get() string {
	if s.FromFile != "" {
		// File contents tend to end with newline
		return strings.TrimSpace(s.FromFile)
	} else {
		return s.Value
	}
}

func (s *Secret) IsSet() bool {
	return s.Value != "" || s.FromFile != ""
}

type EmailConfig struct {
	SmtpServer   string `yaml:"smtp_server" env:"SMTP_SERVER"`
	SmtpPort     int    `yaml:"smtp_port" env:"SMTP_PORT"`
	SmtpUsername string `yaml:"smtp_username" env:"SMTP_USERNAME"`
	SmtpPassword Secret `yaml:"smtp_password" envPrefix:"SMTP_PASSWORD"`
	// If true, then explicitly set the network
	// SSL option for the SMTP client.
	SmtpSSL bool   `yaml:"smtp_ssl" env:"smtp_SSL"`
	SendTo  string `yaml:"send_to" env:"SEND_TO"`
	Sender  string `yaml:"sender" env:"SENDER"`
	Prefix  string `yaml:"prefix" env:"PREFIX"`
	// not bool to allow more flexible usage
	Enabled string `yaml:"enabled" env:"ENABLED"`
	// "never", "beacon", "allow"
	TlsInsecure string `yaml:"tls_insecure" env:"TLS_INSECURE"`
}

// Check if all required fields are set
func (emailConf *EmailConfig) IsConfigured() bool {
	return len(emailConf.MissingConfigurationFields()) == 0
}

func (emailConf *EmailConfig) MissingConfigurationFields() []string {
	missing := []string{}

	if emailConf.SmtpServer == "" {
		missing = append(missing, "smtp_server")
	}
	if emailConf.SmtpPort == 0 {
		missing = append(missing, "smtp_port")
	}
	if emailConf.SmtpUsername == "" {
		missing = append(missing, "smtp_username")
	}
	if !emailConf.SmtpPassword.IsSet() {
		missing = append(missing, "smtp_password")
	}
	if emailConf.SendTo == "" {
		missing = append(missing, "send_to")
	}
	if emailConf.Sender == "" {
		missing = append(missing, "sender")
	}

	return missing
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
	return emailConf.IsConfigured()
}

func (s *Secret) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node got %s", node.Value)
	}
	s.Value = node.Value
	return nil
}

func (s Secret) MarshalYAML() (interface{}, error) {
	return "*****", nil
}

// TzLocation wraps a *time.Location so we can provide custom YAML unmarshalling.
type TzLocation struct {
	Location *time.Location
}

type Config struct {
	Timezone TzLocation `yaml:"timezone" env:"TIMEZONE"`
	// report after n-th hour in the day
	// e.g. 17 -> report after 5pm
	ReportAfter     int           `yaml:"report_time" env:"REPORT_TIME"`
	ReportOnDays    WeekdaysSet   `yaml:"report_on_days" env:"REPORT_ON_DAYS"`
	DbPath          string        `yaml:"db_path" env:"DB"`
	ReportName      string        `yaml:"report_name" env:"REPORT_NAME"`
	Port            int           `yaml:"port" env:"PORT"`
	SchedulerPeriod time.Duration `yaml:"scheduler_period" env:"SCHEDULER_PERIOD"`
	WebCheckPeriod  time.Duration `yaml:"web_check_period" env:"WEB_CHECK_PERIOD"`

	EmailConf EmailConfig `yaml:"email" envPrefix:"EMAIL_"`

	Services ServicesList

	AllowUnknownHeartbeats bool
	RequireHeartbeatAuth   bool

	// RabbitMQ connection string, "amqp://guest:guest@localhost:5672/"
	RabbitConn string `yaml:"rabbit_conn" env:"RABBIT_CONN"`

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

//go:embed config.default.yaml
var DEFAULT_CONFIG_FILE []byte

func ensureConfigFile(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.WriteFile(path, DEFAULT_CONFIG_FILE, 0644)
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

// Default config
func NewConfig() *Config {
	config := &Config{
		Timezone:               TzLocation{time.Local},
		ReportAfter:            17,
		Port:                   8088,
		SchedulerPeriod:        15 * time.Minute,
		WebCheckPeriod:         15 * time.Minute,
		AllowUnknownHeartbeats: true,
		RequireHeartbeatAuth:   false,
		envPrefix:              ENV_VAR_PREFIX,
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
	logger := logging.Get()
	config := NewConfig()
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	err = env.ParseWithOptions(config, env.Options{
		Prefix: ENV_VAR_PREFIX,
	})
	if err != nil {
		return nil, err
	}
	logger.Infow("loaded config", "config", config)
	return config, err
}

func configFromFile(configFile string) (*Config, error) {
	logger := logging.Get()
	logger.Infow("reading config from file", "path", configFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return ConfigFromBytes(data)
}
