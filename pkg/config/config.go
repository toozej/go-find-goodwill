package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration structure.
type Config struct {
	// ShopGoodwill configuration
	ShopGoodwill ShopGoodwillConfig `yaml:"shopgoodwill"`

	// Search configuration
	Search SearchConfig `yaml:"search"`

	// Notification configuration
	Notification NotificationConfig `yaml:"notification"`

	// Web server configuration
	Web WebConfig `yaml:"web"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// Anti-bot configuration
	AntiBot AntiBotConfig `yaml:"antibot"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`
}

// ShopGoodwillConfig contains ShopGoodwill API configuration
type ShopGoodwillConfig struct {
	Username       string        `env:"GOODWILL_USERNAME" yaml:"username"`
	Password       string        `env:"GOODWILL_PASSWORD" yaml:"password"`
	APIBaseURL     string        `env:"GOODWILL_API_BASE_URL" envDefault:"https://api.shopgoodwill.com" yaml:"api_base_url"`
	MaxRetries     int           `env:"GOODWILL_MAX_RETRIES" envDefault:"3" yaml:"max_retries"`
	RequestTimeout time.Duration `env:"GOODWILL_REQUEST_TIMEOUT" envDefault:"30s" yaml:"request_timeout"`
}

// SearchConfig contains search-related configuration
type SearchConfig struct {
	IntervalMinutes           int  `env:"SEARCH_INTERVAL_MINUTES" envDefault:"15" yaml:"interval_minutes"`
	MinInterval               int  `env:"SEARCH_MIN_INTERVAL" envDefault:"5" yaml:"min_interval"`
	MaxInterval               int  `env:"SEARCH_MAX_INTERVAL" envDefault:"1440" yaml:"max_interval"`
	NotificationThresholdDays int  `env:"NOTIFICATION_THRESHOLD_DAYS" envDefault:"1" yaml:"notification_threshold_days"`
	EnableRegexSearches       bool `env:"ENABLE_REGEX_SEARCHES" envDefault:"false" yaml:"enable_regex_searches"`
	Concurrency               int  `env:"SEARCH_CONCURRENCY" envDefault:"3" yaml:"concurrency"`
}

// NotificationConfig contains notification configuration
type NotificationConfig struct {
	Gotify     GotifyConfig     `yaml:"gotify"`
	Slack      SlackConfig      `yaml:"slack"`
	Telegram   TelegramConfig   `yaml:"telegram"`
	Discord    DiscordConfig    `yaml:"discord"`
	Pushover   PushoverConfig   `yaml:"pushover"`
	Pushbullet PushbulletConfig `yaml:"pushbullet"`
}

// GotifyConfig contains Gotify notification configuration
type GotifyConfig struct {
	Enabled  bool   `env:"NOTIFICATION_GOTIFY_ENABLED" envDefault:"false" yaml:"enabled"`
	URL      string `env:"NOTIFICATION_GOTIFY_URL" yaml:"url"`
	Token    string `env:"NOTIFICATION_GOTIFY_TOKEN" yaml:"token"`
	Priority int    `env:"NOTIFICATION_GOTIFY_PRIORITY" envDefault:"5" yaml:"priority"`
}

// SlackConfig contains Slack notification configuration
type SlackConfig struct {
	Enabled   bool   `env:"NOTIFICATION_SLACK_ENABLED" envDefault:"false" yaml:"enabled"`
	Token     string `env:"NOTIFICATION_SLACK_TOKEN" yaml:"token"`
	ChannelID string `env:"NOTIFICATION_SLACK_CHANNEL_ID" yaml:"channel_id"`
}

// TelegramConfig contains Telegram notification configuration
type TelegramConfig struct {
	Enabled bool   `env:"NOTIFICATION_TELEGRAM_ENABLED" envDefault:"false" yaml:"enabled"`
	Token   string `env:"NOTIFICATION_TELEGRAM_TOKEN" yaml:"token"`
	ChatID  string `env:"NOTIFICATION_TELEGRAM_CHAT_ID" yaml:"chat_id"`
}

// DiscordConfig contains Discord notification configuration
type DiscordConfig struct {
	Enabled   bool   `env:"NOTIFICATION_DISCORD_ENABLED" envDefault:"false" yaml:"enabled"`
	Token     string `env:"NOTIFICATION_DISCORD_TOKEN" yaml:"token"`
	ChannelID string `env:"NOTIFICATION_DISCORD_CHANNEL_ID" yaml:"channel_id"`
}

// PushoverConfig contains Pushover notification configuration
type PushoverConfig struct {
	Enabled     bool   `env:"NOTIFICATION_PUSHOVER_ENABLED" envDefault:"false" yaml:"enabled"`
	Token       string `env:"NOTIFICATION_PUSHOVER_TOKEN" yaml:"token"`
	RecipientID string `env:"NOTIFICATION_PUSHOVER_RECIPIENT_ID" yaml:"recipient_id"`
}

// PushbulletConfig contains Pushbullet notification configuration
type PushbulletConfig struct {
	Enabled        bool   `env:"NOTIFICATION_PUSHBULLET_ENABLED" envDefault:"false" yaml:"enabled"`
	Token          string `env:"NOTIFICATION_PUSHBULLET_TOKEN" yaml:"token"`
	DeviceNickname string `env:"NOTIFICATION_PUSHBULLET_DEVICE_NICKNAME" yaml:"device_nickname"`
}

// WebConfig contains web server configuration
type WebConfig struct {
	Enabled      bool          `env:"WEB_SERVER_ENABLED" envDefault:"false" yaml:"enabled"`
	Host         string        `env:"WEB_SERVER_HOST" envDefault:"0.0.0.0" yaml:"host"`
	Port         int           `env:"WEB_SERVER_PORT" envDefault:"8080" yaml:"port"`
	TLS          TLSConfig     `yaml:"tls"`
	StaticDir    string        `env:"WEB_SERVER_STATIC_DIR" envDefault:"web/static" yaml:"static_dir"`
	TemplateDir  string        `env:"WEB_SERVER_TEMPLATE_DIR" envDefault:"web/templates" yaml:"template_dir"`
	ReadTimeout  time.Duration `env:"WEB_SERVER_READ_TIMEOUT" envDefault:"30s" yaml:"read_timeout"`
	WriteTimeout time.Duration `env:"WEB_SERVER_WRITE_TIMEOUT" envDefault:"30s" yaml:"write_timeout"`
	IdleTimeout  time.Duration `env:"WEB_SERVER_IDLE_TIMEOUT" envDefault:"120s" yaml:"idle_timeout"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled  bool   `env:"WEB_SERVER_TLS_ENABLED" envDefault:"false" yaml:"enabled"`
	CertFile string `env:"WEB_SERVER_TLS_CERT" yaml:"cert_file"`
	KeyFile  string `env:"WEB_SERVER_TLS_KEY" yaml:"key_file"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Path               string        `env:"DB_PATH" envDefault:"./goodwill.db" yaml:"path"`
	MaxConnections     int           `env:"DB_MAX_CONNECTIONS" envDefault:"10" yaml:"max_connections"`
	MaxIdleConnections int           `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"5" yaml:"max_idle_connections"`
	ConnectionTimeout  time.Duration `env:"DB_CONNECTION_TIMEOUT" envDefault:"30s" yaml:"connection_timeout"`
	ConnMaxLifetime    time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"10m" yaml:"conn_max_idle_time"`
}

// AntiBotConfig contains anti-bot configuration
type AntiBotConfig struct {
	UserAgent  UserAgentConfig  `yaml:"user_agent"`
	Timing     TimingConfig     `yaml:"timing"`
	Throttling ThrottlingConfig `yaml:"throttling"`
	Retry      RetryConfig      `yaml:"retry"`
	Circuit    CircuitConfig    `yaml:"circuit"`
}

// UserAgentConfig contains user agent rotation configuration
type UserAgentConfig struct {
	RotationEnabled  bool          `env:"ANTIBOT_USER_AGENT_ROTATION" envDefault:"false" yaml:"rotation_enabled"`
	RotationInterval time.Duration `env:"ANTIBOT_ROTATION_INTERVAL" envDefault:"1h" yaml:"rotation_interval"`
	RequestsPerUA    int           `env:"ANTIBOT_REQUESTS_PER_UA" envDefault:"20" yaml:"requests_per_ua"`
	MinSuccessRate   float64       `env:"ANTIBOT_MIN_SUCCESS_RATE" envDefault:"0.8" yaml:"min_success_rate"`
}

// TimingConfig contains timing configuration for anti-bot measures
type TimingConfig struct {
	BaseInterval       time.Duration `env:"ANTIBOT_BASE_INTERVAL" envDefault:"15m" yaml:"base_interval"`
	MinJitter          time.Duration `env:"ANTIBOT_MIN_JITTER" envDefault:"2m" yaml:"min_jitter"`
	MaxJitter          time.Duration `env:"ANTIBOT_MAX_JITTER" envDefault:"5m" yaml:"max_jitter"`
	HumanLikeVariation bool          `env:"ANTIBOT_HUMAN_VARIATION" envDefault:"false" yaml:"human_like_variation"`
	MaxQueueSize       int           `env:"ANTIBOT_MAX_QUEUE_SIZE" envDefault:"100" yaml:"max_queue_size"`
}

// ThrottlingConfig contains request throttling configuration
type ThrottlingConfig struct {
	RequestsPerMinute int `env:"ANTIBOT_REQUESTS_PER_MINUTE" envDefault:"0" yaml:"requests_per_minute"`
	BurstLimit        int `env:"ANTIBOT_BURST_LIMIT" envDefault:"0" yaml:"burst_limit"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxRetries int           `env:"ANTIBOT_RETRY_MAX_RETRIES" envDefault:"3" yaml:"max_retries"`
	BaseDelay  time.Duration `env:"ANTIBOT_RETRY_BASE_DELAY" envDefault:"1s" yaml:"base_delay"`
	MaxDelay   time.Duration `env:"ANTIBOT_RETRY_MAX_DELAY" envDefault:"30s" yaml:"max_delay"`
}

// CircuitConfig contains circuit breaker configuration
type CircuitConfig struct {
	FailureThreshold int           `env:"ANTIBOT_CIRCUIT_FAILURE_THRESHOLD" envDefault:"3" yaml:"failure_threshold"`
	SuccessThreshold int           `env:"ANTIBOT_CIRCUIT_SUCCESS_THRESHOLD" envDefault:"2" yaml:"success_threshold"`
	Timeout          time.Duration `env:"ANTIBOT_CIRCUIT_TIMEOUT" envDefault:"30s" yaml:"timeout"`
	ResetTimeout     time.Duration `env:"ANTIBOT_CIRCUIT_RESET_TIMEOUT" envDefault:"5m" yaml:"reset_timeout"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `env:"LOG_LEVEL" envDefault:"info" yaml:"level"`
	Format     string `env:"LOG_FORMAT" envDefault:"text" yaml:"format"`
	File       string `env:"LOG_FILE" yaml:"file"`
	MaxSize    int    `env:"LOG_MAX_SIZE" envDefault:"10" yaml:"max_size"`
	MaxBackups int    `env:"LOG_MAX_BACKUPS" envDefault:"3" yaml:"max_backups"`
	MaxAge     int    `env:"LOG_MAX_AGE" envDefault:"7" yaml:"max_age"`
}

// GetEnvVars loads and returns the application configuration from environment
// variables and .env files with comprehensive security validation.
func GetEnvVars() (Config, error) {
	// Get current working directory for secure file operations
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("error getting current working directory: %w", err)
	}

	// Construct path for .env file within current directory
	envPath := filepath.Join(cwd, ".env")

	// Start with current process environment variables
	environ := os.Environ()
	envMap := make(map[string]string)
	for _, e := range environ {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}

	// Load .env file if it exists, without modifying process environment
	if _, err := os.Stat(envPath); err == nil {
		updates, err := godotenv.Read(envPath)
		if err != nil {
			return Config{}, fmt.Errorf("error reading .env file: %w", err)
		}
		// Merge .env values (do not overwrite existing env vars?)
		// godotenv.Load does NOT overwrite by default.
		// So we only add if not present.
		for k, v := range updates {
			if _, exists := envMap[k]; !exists {
				envMap[k] = v
			}
		}
	}

	// Parse environment variables into config struct
	var conf Config
	// Parse with options using our constructed map
	if err := env.ParseWithOptions(&conf, env.Options{Environment: envMap}); err != nil {
		return Config{}, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return conf, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate ShopGoodwill configuration
	if c.ShopGoodwill.Username == "" {
		return fmt.Errorf("shopgoodwill username is required")
	}
	if c.ShopGoodwill.Password == "" {
		return fmt.Errorf("shopgoodwill password is required")
	}

	// Validate search configuration
	if c.Search.IntervalMinutes < c.Search.MinInterval {
		return fmt.Errorf("search interval cannot be less than min interval")
	}
	if c.Search.IntervalMinutes > c.Search.MaxInterval {
		return fmt.Errorf("search interval cannot be greater than max interval")
	}

	// Validate notification configuration
	if c.Notification.Gotify.Enabled && c.Notification.Gotify.URL == "" {
		return fmt.Errorf("gotify URL is required when gotify is enabled")
	}
	if c.Notification.Slack.Enabled && c.Notification.Slack.Token == "" {
		return fmt.Errorf("slack token is required when slack is enabled")
	}
	if c.Notification.Telegram.Enabled && c.Notification.Telegram.Token == "" {
		return fmt.Errorf("telegram token is required when telegram is enabled")
	}
	if c.Notification.Discord.Enabled && c.Notification.Discord.Token == "" {
		return fmt.Errorf("discord token is required when discord is enabled")
	}
	if c.Notification.Pushover.Enabled && c.Notification.Pushover.Token == "" {
		return fmt.Errorf("pushover token is required when pushover is enabled")
	}
	if c.Notification.Pushbullet.Enabled && c.Notification.Pushbullet.Token == "" {
		return fmt.Errorf("pushbullet token is required when pushbullet is enabled")
	}

	// Validate web server configuration
	if c.Web.TLS.Enabled {
		if c.Web.TLS.CertFile == "" {
			return fmt.Errorf("TLS certificate file is required when TLS is enabled")
		}
		if c.Web.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	return nil
}

// LoadYAMLConfig loads configuration from a YAML file and merges it with environment variables
func LoadYAMLConfig(yamlPath string) (Config, error) {
	// Validate the YAML file path
	if yamlPath == "" {
		return Config{}, fmt.Errorf("YAML file path cannot be empty")
	}

	path, err := filepath.Abs(yamlPath)
	if err != nil {
		return Config{}, fmt.Errorf("error resolving YAML file path: %w", err)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return Config{}, fmt.Errorf("YAML file must have .yaml or .yml extension")
	}

	// Read and parse YAML
	// We allow reading from any path provided by the caller/user
	cleanPath := filepath.Clean(path)
	yamlData, err := os.ReadFile(cleanPath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading YAML file: %w", err)
	}

	var conf Config
	// 1. Load Defaults and Environment Variables
	// Use GetEnvVars to avoid side effects
	envConf, err := GetEnvVars()
	if err != nil {
		return Config{}, fmt.Errorf("error loading environment variables: %w", err)
	}
	conf = envConf

	// 2. Load YAML
	var yamlConf Config
	if err := yaml.Unmarshal(yamlData, &yamlConf); err != nil {
		return Config{}, fmt.Errorf("error parsing YAML configuration: %w", err)
	}

	// 3. Merge YAML into conf (YAML overrides Defaults, but Env overrides YAML)
	if err := mergeConfigStructs(&conf, &yamlConf); err != nil {
		return Config{}, fmt.Errorf("error merging configurations: %w", err)
	}

	return conf, nil
}

// mergeConfigStructs merges yamlConf into conf, respecting Env > YAML > Defaults precedence.
// conf contains (Defaults + Env). yamlConf contains YAML.
// We overwrite conf with yamlConf value ONLY if:
// 1. yamlConf value is non-zero (present in YAML)
// 2. AND the corresponding Env Var is NOT set (so conf has Default, which YAML should override).
func mergeConfigStructs(conf, yamlConf interface{}) error {
	vConf := reflect.ValueOf(conf).Elem()
	vYaml := reflect.ValueOf(yamlConf).Elem()
	tConf := vConf.Type()

	for i := 0; i < vConf.NumField(); i++ {
		fieldConf := vConf.Field(i)
		fieldYaml := vYaml.Field(i)
		fieldType := tConf.Field(i)

		// Handle nested structs recursively
		if fieldConf.Kind() == reflect.Struct {
			if err := mergeConfigStructs(fieldConf.Addr().Interface(), fieldYaml.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// Get environment variable name from tag
		envTag := fieldType.Tag.Get("env")
		// Clean tag options (e.g. "NAME,required")
		envVar := strings.Split(envTag, ",")[0]

		// Check if YAML has a value
		if !fieldYaml.IsZero() {
			// Check if Env Var is set
			envSet := false
			if envVar != "" {
				_, envSet = os.LookupEnv(envVar)
			}

			// If Env Var is NOT set, YAML wins (overrides Default)
			if !envSet {
				fieldConf.Set(fieldYaml)
			}
		}
	}
	return nil
}
