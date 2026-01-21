package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetEnvVars(t *testing.T) {
	tests := []struct {
		name               string
		mockEnv            map[string]string
		mockEnvFile        string
		expectError        bool
		expectedConfig     Config
		validationExpected bool
	}{
		{
			name: "Valid environment variables with full config",
			mockEnv: map[string]string{
				"GOODWILL_USERNAME":            "testuser",
				"GOODWILL_PASSWORD":            "testpass",
				"GOODWILL_API_BASE_URL":        "https://test-api.shopgoodwill.com",
				"GOODWILL_MAX_RETRIES":         "5",
				"GOODWILL_REQUEST_TIMEOUT":     "60s",
				"SEARCH_INTERVAL_MINUTES":      "30",
				"SEARCH_MIN_INTERVAL":          "10",
				"SEARCH_MAX_INTERVAL":          "2880",
				"NOTIFICATION_THRESHOLD_DAYS":  "3",
				"ENABLE_REGEX_SEARCHES":        "true",
				"SEARCH_CONCURRENCY":           "5",
				"NOTIFICATION_GOTIFY_ENABLED":  "true",
				"NOTIFICATION_GOTIFY_URL":      "https://gotify.example.com",
				"NOTIFICATION_GOTIFY_TOKEN":    "gotifytoken",
				"NOTIFICATION_GOTIFY_PRIORITY": "7",
				"WEB_SERVER_ENABLED":           "true",
				"WEB_SERVER_HOST":              "127.0.0.1",
				"WEB_SERVER_PORT":              "9090",
				"WEB_SERVER_TLS_ENABLED":       "true",
				"WEB_SERVER_TLS_CERT":          "/path/to/cert.pem",
				"WEB_SERVER_TLS_KEY":           "/path/to/key.pem",
				"WEB_SERVER_STATIC_DIR":        "./static",
				"WEB_SERVER_TEMPLATE_DIR":      "./templates",
				"WEB_SERVER_READ_TIMEOUT":      "60s",
				"WEB_SERVER_WRITE_TIMEOUT":     "60s",
				"WEB_SERVER_IDLE_TIMEOUT":      "240s",
				"DB_PATH":                      "./test.db",
				"DB_MAX_CONNECTIONS":           "20",
				"DB_CONNECTION_TIMEOUT":        "60s",
				"ANTIBOT_USER_AGENT_ROTATION":  "true",
				"ANTIBOT_ROTATION_INTERVAL":    "30m",
				"ANTIBOT_REQUESTS_PER_UA":      "50",
				"ANTIBOT_MIN_SUCCESS_RATE":     "0.9",
				"ANTIBOT_BASE_INTERVAL":        "30m",
				"ANTIBOT_MIN_JITTER":           "5m",
				"ANTIBOT_MAX_JITTER":           "10m",
				"ANTIBOT_HUMAN_VARIATION":      "true",
				"ANTIBOT_REQUESTS_PER_MINUTE":  "100",
				"ANTIBOT_BURST_LIMIT":          "20",
				"LOG_LEVEL":                    "debug",
				"LOG_FORMAT":                   "json",
				"LOG_FILE":                     "./app.log",
				"LOG_MAX_SIZE":                 "20",
				"LOG_MAX_BACKUPS":              "5",
				"LOG_MAX_AGE":                  "14",
			},
			expectedConfig: Config{
				ShopGoodwill: ShopGoodwillConfig{
					Username:       "testuser",
					Password:       "testpass",
					APIBaseURL:     "https://test-api.shopgoodwill.com",
					MaxRetries:     5,
					RequestTimeout: 60 * time.Second,
				},
				Search: SearchConfig{
					IntervalMinutes:           30,
					MinInterval:               10,
					MaxInterval:               2880,
					NotificationThresholdDays: 3,
					EnableRegexSearches:       true,
					Concurrency:               5,
				},
				Notification: NotificationConfig{
					Gotify: GotifyConfig{
						Enabled:  true,
						URL:      "https://gotify.example.com",
						Token:    "gotifytoken",
						Priority: 7,
					},
				},
				Web: WebConfig{
					Enabled: true,
					Host:    "127.0.0.1",
					Port:    9090,
					TLS: TLSConfig{
						Enabled:  true,
						CertFile: "/path/to/cert.pem",
						KeyFile:  "/path/to/key.pem",
					},
					StaticDir:    "./static",
					TemplateDir:  "./templates",
					ReadTimeout:  60 * time.Second,
					WriteTimeout: 60 * time.Second,
					IdleTimeout:  240 * time.Second,
				},
				Database: DatabaseConfig{
					Path:              "./test.db",
					MaxConnections:    20,
					ConnectionTimeout: 60 * time.Second,
				},
				AntiBot: AntiBotConfig{
					UserAgent: UserAgentConfig{
						RotationEnabled:  true,
						RotationInterval: 30 * time.Minute,
						RequestsPerUA:    50,
						MinSuccessRate:   0.9,
					},
					Timing: TimingConfig{
						BaseInterval:       30 * time.Minute,
						MinJitter:          5 * time.Minute,
						MaxJitter:          10 * time.Minute,
						HumanLikeVariation: true,
					},
					Throttling: ThrottlingConfig{
						RequestsPerMinute: 100,
						BurstLimit:        20,
					},
				},
				Logging: LoggingConfig{
					Level:      "debug",
					Format:     "json",
					File:       "./app.log",
					MaxSize:    20,
					MaxBackups: 5,
					MaxAge:     14,
				},
			},
			validationExpected: true,
		},
		{
			name: "Valid .env file with full config",
			mockEnvFile: `GOODWILL_USERNAME=testenvfileuser
GOODWILL_PASSWORD=envfilepass
GOODWILL_API_BASE_URL=https://envfile-api.shopgoodwill.com
GOODWILL_MAX_RETRIES=4
GOODWILL_REQUEST_TIMEOUT=45s
SEARCH_INTERVAL_MINUTES=20
SEARCH_MIN_INTERVAL=8
SEARCH_MAX_INTERVAL=2000
NOTIFICATION_THRESHOLD_DAYS=2
ENABLE_REGEX_SEARCHES=false
SEARCH_CONCURRENCY=4
NOTIFICATION_GOTIFY_ENABLED=true
NOTIFICATION_GOTIFY_URL=https://gotify-env.example.com
NOTIFICATION_GOTIFY_TOKEN=envgotifytoken
NOTIFICATION_GOTIFY_PRIORITY=6
WEB_SERVER_ENABLED=true
WEB_SERVER_HOST=0.0.0.0
WEB_SERVER_PORT=8081
WEB_SERVER_TLS_ENABLED=true
WEB_SERVER_TLS_CERT=/env/path/to/cert.pem
WEB_SERVER_TLS_KEY=/env/path/to/key.pem
WEB_SERVER_STATIC_DIR=./envstatic
WEB_SERVER_TEMPLATE_DIR=./envtemplates
WEB_SERVER_READ_TIMEOUT=45s
WEB_SERVER_WRITE_TIMEOUT=45s
WEB_SERVER_IDLE_TIMEOUT=180s
DB_PATH=./env-test.db
DB_MAX_CONNECTIONS=15
DB_CONNECTION_TIMEOUT=45s
ANTIBOT_USER_AGENT_ROTATION=true
ANTIBOT_ROTATION_INTERVAL=20m
ANTIBOT_REQUESTS_PER_UA=30
ANTIBOT_MIN_SUCCESS_RATE=0.85
ANTIBOT_BASE_INTERVAL=25m
ANTIBOT_MIN_JITTER=4m
ANTIBOT_MAX_JITTER=8m
ANTIBOT_HUMAN_VARIATION=true
ANTIBOT_REQUESTS_PER_MINUTE=80
ANTIBOT_BURST_LIMIT=15
LOG_LEVEL=warn
LOG_FORMAT=text
LOG_FILE=./env-app.log
LOG_MAX_SIZE=15
LOG_MAX_BACKUPS=4
LOG_MAX_AGE=10`,
			expectedConfig: Config{
				ShopGoodwill: ShopGoodwillConfig{
					Username:       "testenvfileuser",
					Password:       "envfilepass",
					APIBaseURL:     "https://envfile-api.shopgoodwill.com",
					MaxRetries:     4,
					RequestTimeout: 45 * time.Second,
				},
				Search: SearchConfig{
					IntervalMinutes:           20,
					MinInterval:               8,
					MaxInterval:               2000,
					NotificationThresholdDays: 2,
					EnableRegexSearches:       false,
					Concurrency:               4,
				},
				Notification: NotificationConfig{
					Gotify: GotifyConfig{
						Enabled:  true,
						URL:      "https://gotify-env.example.com",
						Token:    "envgotifytoken",
						Priority: 6,
					},
				},
				Web: WebConfig{
					Enabled: true,
					Host:    "0.0.0.0",
					Port:    8081,
					TLS: TLSConfig{
						Enabled:  true,
						CertFile: "/env/path/to/cert.pem",
						KeyFile:  "/env/path/to/key.pem",
					},
					StaticDir:    "./envstatic",
					TemplateDir:  "./envtemplates",
					ReadTimeout:  45 * time.Second,
					WriteTimeout: 45 * time.Second,
					IdleTimeout:  180 * time.Second,
				},
				Database: DatabaseConfig{
					Path:              "./env-test.db",
					MaxConnections:    15,
					ConnectionTimeout: 45 * time.Second,
				},
				AntiBot: AntiBotConfig{
					UserAgent: UserAgentConfig{
						RotationEnabled:  true,
						RotationInterval: 20 * time.Minute,
						RequestsPerUA:    30,
						MinSuccessRate:   0.85,
					},
					Timing: TimingConfig{
						BaseInterval:       25 * time.Minute,
						MinJitter:          4 * time.Minute,
						MaxJitter:          8 * time.Minute,
						HumanLikeVariation: true,
					},
					Throttling: ThrottlingConfig{
						RequestsPerMinute: 80,
						BurstLimit:        15,
					},
				},
				Logging: LoggingConfig{
					Level:      "warn",
					Format:     "text",
					File:       "./env-app.log",
					MaxSize:    15,
					MaxBackups: 4,
					MaxAge:     10,
				},
			},
			validationExpected: true,
		},
		{
			name: "Environment variable overrides .env file",
			mockEnv: map[string]string{
				"GOODWILL_USERNAME": "envuser",
				"GOODWILL_PASSWORD": "envpass",
			},
			mockEnvFile: `GOODWILL_USERNAME=fileuser
GOODWILL_PASSWORD=filepass`,
			expectedConfig: Config{
				ShopGoodwill: ShopGoodwillConfig{
					Username:       "envuser",
					Password:       "envpass",
					APIBaseURL:     "https://api.shopgoodwill.com", // defaults
					MaxRetries:     3,                              // defaults
					RequestTimeout: 30 * time.Second,               // defaults
				},
				Search: SearchConfig{
					IntervalMinutes:           15,    // defaults
					MinInterval:               5,     // defaults
					MaxInterval:               1440,  // defaults
					NotificationThresholdDays: 1,     // defaults
					EnableRegexSearches:       false, // defaults
					Concurrency:               3,     // defaults
				},
				Notification: NotificationConfig{
					Gotify: GotifyConfig{
						Enabled:  false, // defaults
						URL:      "",    // defaults
						Token:    "",    // defaults
						Priority: 5,     // defaults
					},
				},
				Web: WebConfig{
					Enabled: false,     // defaults
					Host:    "0.0.0.0", // defaults
					Port:    8080,      // defaults
					TLS: TLSConfig{
						Enabled:  false, // defaults
						CertFile: "",    // defaults
						KeyFile:  "",    // defaults
					},
					StaticDir:    "web/static",      // defaults
					TemplateDir:  "web/templates",   // defaults
					ReadTimeout:  30 * time.Second,  // defaults
					WriteTimeout: 30 * time.Second,  // defaults
					IdleTimeout:  120 * time.Second, // defaults
				},
				Database: DatabaseConfig{
					Path:              "./goodwill.db",  // defaults
					MaxConnections:    10,               // defaults
					ConnectionTimeout: 30 * time.Second, // defaults
				},
				AntiBot: AntiBotConfig{
					UserAgent: UserAgentConfig{
						RotationEnabled:  false,     // defaults
						RotationInterval: time.Hour, // defaults
						RequestsPerUA:    20,        // defaults
						MinSuccessRate:   0.8,       // defaults
					},
					Timing: TimingConfig{
						BaseInterval:       15 * time.Minute, // defaults
						MinJitter:          2 * time.Minute,  // defaults
						MaxJitter:          5 * time.Minute,  // defaults
						HumanLikeVariation: false,            // defaults
					},
					Throttling: ThrottlingConfig{
						RequestsPerMinute: 0, // defaults
						BurstLimit:        0, // defaults
					},
				},
				Logging: LoggingConfig{
					Level:      "info", // defaults
					Format:     "text", // defaults
					File:       "",     // defaults
					MaxSize:    10,     // defaults
					MaxBackups: 3,      // defaults
					MaxAge:     7,      // defaults
				},
			},
			validationExpected: false,
		},
		{
			name: "No environment variables or .env file - defaults applied",
			expectedConfig: Config{
				ShopGoodwill: ShopGoodwillConfig{
					Username:       "",
					Password:       "",
					APIBaseURL:     "https://api.shopgoodwill.com",
					MaxRetries:     3,
					RequestTimeout: 30 * time.Second,
				},
				Search: SearchConfig{
					IntervalMinutes:           15,
					MinInterval:               5,
					MaxInterval:               1440,
					NotificationThresholdDays: 1,
					EnableRegexSearches:       false,
					Concurrency:               3,
				},
				Notification: NotificationConfig{
					Gotify: GotifyConfig{
						Enabled:  false,
						URL:      "",
						Token:    "",
						Priority: 5,
					},
				},
				Web: WebConfig{
					Enabled: false,
					Host:    "0.0.0.0",
					Port:    8080,
					TLS: TLSConfig{
						Enabled:  false,
						CertFile: "",
						KeyFile:  "",
					},
					StaticDir:    "web/static",
					TemplateDir:  "web/templates",
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
				Database: DatabaseConfig{
					Path:              "./goodwill.db",
					MaxConnections:    10,
					ConnectionTimeout: 30 * time.Second,
				},
				AntiBot: AntiBotConfig{
					UserAgent: UserAgentConfig{
						RotationEnabled:  false,
						RotationInterval: time.Hour,
						RequestsPerUA:    20,
						MinSuccessRate:   0.8,
					},
					Timing: TimingConfig{
						BaseInterval:       15 * time.Minute,
						MinJitter:          2 * time.Minute,
						MaxJitter:          5 * time.Minute,
						HumanLikeVariation: false,
					},
					Throttling: ThrottlingConfig{
						RequestsPerMinute: 0,
						BurstLimit:        0,
					},
				},
				Logging: LoggingConfig{
					Level:      "info",
					Format:     "text",
					File:       "",
					MaxSize:    10,
					MaxBackups: 3,
					MaxAge:     7,
				},
			},
			validationExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original directory and change to temp directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}

			tmpDir := t.TempDir()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("Failed to restore original directory: %v", err)
				}
			}()

			// Create .env file if applicable
			if tt.mockEnvFile != "" {
				envPath := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envPath, []byte(tt.mockEnvFile), 0644); err != nil {
					t.Fatalf("Failed to write mock .env file: %v", err)
				}
			}

			// Clear all environment variables that might affect the test
			envVarsToClear := []string{
				"GOODWILL_USERNAME", "GOODWILL_PASSWORD", "GOODWILL_API_BASE_URL",
				"GOODWILL_MAX_RETRIES", "GOODWILL_REQUEST_TIMEOUT",
				"SEARCH_INTERVAL_MINUTES", "SEARCH_MIN_INTERVAL", "SEARCH_MAX_INTERVAL",
				"NOTIFICATION_THRESHOLD_DAYS", "ENABLE_REGEX_SEARCHES", "SEARCH_CONCURRENCY",
				"NOTIFICATION_GOTIFY_ENABLED", "NOTIFICATION_GOTIFY_URL", "NOTIFICATION_GOTIFY_TOKEN",
				"NOTIFICATION_GOTIFY_PRIORITY", "WEB_SERVER_ENABLED", "WEB_SERVER_HOST",
				"WEB_SERVER_PORT", "WEB_SERVER_TLS_ENABLED", "WEB_SERVER_TLS_CERT",
				"WEB_SERVER_TLS_KEY", "WEB_SERVER_STATIC_DIR", "WEB_SERVER_TEMPLATE_DIR",
				"WEB_SERVER_READ_TIMEOUT", "WEB_SERVER_WRITE_TIMEOUT", "WEB_SERVER_IDLE_TIMEOUT",
				"DB_PATH", "DB_MAX_CONNECTIONS", "DB_CONNECTION_TIMEOUT",
				"ANTIBOT_USER_AGENT_ROTATION", "ANTIBOT_ROTATION_INTERVAL",
				"ANTIBOT_REQUESTS_PER_UA", "ANTIBOT_MIN_SUCCESS_RATE", "ANTIBOT_BASE_INTERVAL",
				"ANTIBOT_MIN_JITTER", "ANTIBOT_MAX_JITTER", "ANTIBOT_HUMAN_VARIATION",
				"ANTIBOT_REQUESTS_PER_MINUTE", "ANTIBOT_BURST_LIMIT", "LOG_LEVEL",
				"LOG_FORMAT", "LOG_FILE", "LOG_MAX_SIZE", "LOG_MAX_BACKUPS", "LOG_MAX_AGE",
			}

			for _, key := range envVarsToClear {
				os.Unsetenv(key)
			}

			// Set mock environment variables (these should override .env file)
			for key, value := range tt.mockEnv {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.mockEnv {
					os.Unsetenv(key)
				}
			}()

			// Call function
			conf, err := GetEnvVars()
			if err != nil {
				t.Fatalf("GetEnvVars() returned unexpected error: %v", err)
			}

			// Verify ShopGoodwill configuration
			if conf.ShopGoodwill.Username != tt.expectedConfig.ShopGoodwill.Username {
				t.Errorf("ShopGoodwill.Username: expected %q, got %q", tt.expectedConfig.ShopGoodwill.Username, conf.ShopGoodwill.Username)
			}
			if conf.ShopGoodwill.Password != tt.expectedConfig.ShopGoodwill.Password {
				t.Errorf("ShopGoodwill.Password: expected %q, got %q", tt.expectedConfig.ShopGoodwill.Password, conf.ShopGoodwill.Password)
			}
			if conf.ShopGoodwill.APIBaseURL != tt.expectedConfig.ShopGoodwill.APIBaseURL {
				t.Errorf("ShopGoodwill.APIBaseURL: expected %q, got %q", tt.expectedConfig.ShopGoodwill.APIBaseURL, conf.ShopGoodwill.APIBaseURL)
			}
			if conf.ShopGoodwill.MaxRetries != tt.expectedConfig.ShopGoodwill.MaxRetries {
				t.Errorf("ShopGoodwill.MaxRetries: expected %d, got %d", tt.expectedConfig.ShopGoodwill.MaxRetries, conf.ShopGoodwill.MaxRetries)
			}
			if conf.ShopGoodwill.RequestTimeout != tt.expectedConfig.ShopGoodwill.RequestTimeout {
				t.Errorf("ShopGoodwill.RequestTimeout: expected %v, got %v", tt.expectedConfig.ShopGoodwill.RequestTimeout, conf.ShopGoodwill.RequestTimeout)
			}

			// Verify Search configuration
			if conf.Search.IntervalMinutes != tt.expectedConfig.Search.IntervalMinutes {
				t.Errorf("Search.IntervalMinutes: expected %d, got %d", tt.expectedConfig.Search.IntervalMinutes, conf.Search.IntervalMinutes)
			}
			if conf.Search.MinInterval != tt.expectedConfig.Search.MinInterval {
				t.Errorf("Search.MinInterval: expected %d, got %d", tt.expectedConfig.Search.MinInterval, conf.Search.MinInterval)
			}
			if conf.Search.MaxInterval != tt.expectedConfig.Search.MaxInterval {
				t.Errorf("Search.MaxInterval: expected %d, got %d", tt.expectedConfig.Search.MaxInterval, conf.Search.MaxInterval)
			}
			if conf.Search.NotificationThresholdDays != tt.expectedConfig.Search.NotificationThresholdDays {
				t.Errorf("Search.NotificationThresholdDays: expected %d, got %d", tt.expectedConfig.Search.NotificationThresholdDays, conf.Search.NotificationThresholdDays)
			}
			if conf.Search.EnableRegexSearches != tt.expectedConfig.Search.EnableRegexSearches {
				t.Errorf("Search.EnableRegexSearches: expected %t, got %t", tt.expectedConfig.Search.EnableRegexSearches, conf.Search.EnableRegexSearches)
			}
			if conf.Search.Concurrency != tt.expectedConfig.Search.Concurrency {
				t.Errorf("Search.Concurrency: expected %d, got %d", tt.expectedConfig.Search.Concurrency, conf.Search.Concurrency)
			}

			// Verify Notification configuration
			if conf.Notification.Gotify.Enabled != tt.expectedConfig.Notification.Gotify.Enabled {
				t.Errorf("Notification.Gotify.Enabled: expected %t, got %t", tt.expectedConfig.Notification.Gotify.Enabled, conf.Notification.Gotify.Enabled)
			}
			if conf.Notification.Gotify.URL != tt.expectedConfig.Notification.Gotify.URL {
				t.Errorf("Notification.Gotify.URL: expected %q, got %q", tt.expectedConfig.Notification.Gotify.URL, conf.Notification.Gotify.URL)
			}
			if conf.Notification.Gotify.Token != tt.expectedConfig.Notification.Gotify.Token {
				t.Errorf("Notification.Gotify.Token: expected %q, got %q", tt.expectedConfig.Notification.Gotify.Token, conf.Notification.Gotify.Token)
			}
			if conf.Notification.Gotify.Priority != tt.expectedConfig.Notification.Gotify.Priority {
				t.Errorf("Notification.Gotify.Priority: expected %d, got %d", tt.expectedConfig.Notification.Gotify.Priority, conf.Notification.Gotify.Priority)
			}

			// Verify Web configuration
			if conf.Web.Host != tt.expectedConfig.Web.Host {
				t.Errorf("Web.Host: expected %q, got %q", tt.expectedConfig.Web.Host, conf.Web.Host)
			}
			if conf.Web.Port != tt.expectedConfig.Web.Port {
				t.Errorf("Web.Port: expected %d, got %d", tt.expectedConfig.Web.Port, conf.Web.Port)
			}
			if conf.Web.TLS.Enabled != tt.expectedConfig.Web.TLS.Enabled {
				t.Errorf("Web.TLS.Enabled: expected %t, got %t", tt.expectedConfig.Web.TLS.Enabled, conf.Web.TLS.Enabled)
			}
			if conf.Web.TLS.CertFile != tt.expectedConfig.Web.TLS.CertFile {
				t.Errorf("Web.TLS.CertFile: expected %q, got %q", tt.expectedConfig.Web.TLS.CertFile, conf.Web.TLS.CertFile)
			}
			if conf.Web.TLS.KeyFile != tt.expectedConfig.Web.TLS.KeyFile {
				t.Errorf("Web.TLS.KeyFile: expected %q, got %q", tt.expectedConfig.Web.TLS.KeyFile, conf.Web.TLS.KeyFile)
			}

			// Verify Database configuration
			if conf.Database.Path != tt.expectedConfig.Database.Path {
				t.Errorf("Database.Path: expected %q, got %q", tt.expectedConfig.Database.Path, conf.Database.Path)
			}
			if conf.Database.MaxConnections != tt.expectedConfig.Database.MaxConnections {
				t.Errorf("Database.MaxConnections: expected %d, got %d", tt.expectedConfig.Database.MaxConnections, conf.Database.MaxConnections)
			}
			if conf.Database.ConnectionTimeout != tt.expectedConfig.Database.ConnectionTimeout {
				t.Errorf("Database.ConnectionTimeout: expected %v, got %v", tt.expectedConfig.Database.ConnectionTimeout, conf.Database.ConnectionTimeout)
			}

			// Verify AntiBot configuration
			if conf.AntiBot.UserAgent.RotationEnabled != tt.expectedConfig.AntiBot.UserAgent.RotationEnabled {
				t.Errorf("AntiBot.UserAgent.RotationEnabled: expected %t, got %t", tt.expectedConfig.AntiBot.UserAgent.RotationEnabled, conf.AntiBot.UserAgent.RotationEnabled)
			}
			if conf.AntiBot.UserAgent.RotationInterval != tt.expectedConfig.AntiBot.UserAgent.RotationInterval {
				t.Errorf("AntiBot.UserAgent.RotationInterval: expected %v, got %v", tt.expectedConfig.AntiBot.UserAgent.RotationInterval, conf.AntiBot.UserAgent.RotationInterval)
			}
			if conf.AntiBot.UserAgent.RequestsPerUA != tt.expectedConfig.AntiBot.UserAgent.RequestsPerUA {
				t.Errorf("AntiBot.UserAgent.RequestsPerUA: expected %d, got %d", tt.expectedConfig.AntiBot.UserAgent.RequestsPerUA, conf.AntiBot.UserAgent.RequestsPerUA)
			}
			if conf.AntiBot.UserAgent.MinSuccessRate != tt.expectedConfig.AntiBot.UserAgent.MinSuccessRate {
				t.Errorf("AntiBot.UserAgent.MinSuccessRate: expected %f, got %f", tt.expectedConfig.AntiBot.UserAgent.MinSuccessRate, conf.AntiBot.UserAgent.MinSuccessRate)
			}

			// Verify Logging configuration
			if conf.Logging.Level != tt.expectedConfig.Logging.Level {
				t.Errorf("Logging.Level: expected %q, got %q", tt.expectedConfig.Logging.Level, conf.Logging.Level)
			}
			if conf.Logging.Format != tt.expectedConfig.Logging.Format {
				t.Errorf("Logging.Format: expected %q, got %q", tt.expectedConfig.Logging.Format, conf.Logging.Format)
			}
			if conf.Logging.File != tt.expectedConfig.Logging.File {
				t.Errorf("Logging.File: expected %q, got %q", tt.expectedConfig.Logging.File, conf.Logging.File)
			}
			if conf.Logging.MaxSize != tt.expectedConfig.Logging.MaxSize {
				t.Errorf("Logging.MaxSize: expected %d, got %d", tt.expectedConfig.Logging.MaxSize, conf.Logging.MaxSize)
			}
			if conf.Logging.MaxBackups != tt.expectedConfig.Logging.MaxBackups {
				t.Errorf("Logging.MaxBackups: expected %d, got %d", tt.expectedConfig.Logging.MaxBackups, conf.Logging.MaxBackups)
			}
			if conf.Logging.MaxAge != tt.expectedConfig.Logging.MaxAge {
				t.Errorf("Logging.MaxAge: expected %d, got %d", tt.expectedConfig.Logging.MaxAge, conf.Logging.MaxAge)
			}

			// Test validation if expected
			if tt.validationExpected {
				err := conf.Validate()
				if err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}
