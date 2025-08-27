package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	AWS      AWSConfig      `mapstructure:"aws"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	Mode            string        `mapstructure:"mode"` // development, production
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	URI                   string        `mapstructure:"uri"`
	Database              string        `mapstructure:"database"`
	MaxPoolSize           uint64        `mapstructure:"max_pool_size"`
	MinPoolSize           uint64        `mapstructure:"min_pool_size"`
	MaxIdleTime           time.Duration `mapstructure:"max_idle_time"`
	HeartbeatInterval     time.Duration `mapstructure:"heartbeat_interval"`
	RetryWrites           bool          `mapstructure:"retry_writes"`
	ReplicaSet            string        `mapstructure:"replica_set"`
	ServerSelectionTimeout time.Duration `mapstructure:"server_selection_timeout"`
}

type AuthConfig struct {
	JWTSecret             string        `mapstructure:"jwt_secret"`
	JWTExpiration         time.Duration `mapstructure:"jwt_expiration"`
	RefreshTokenExpiration time.Duration `mapstructure:"refresh_token_expiration"`
	PasswordMinLength     int           `mapstructure:"password_min_length"`
	BcryptCost           int           `mapstructure:"bcrypt_cost"`
	MaxLoginAttempts     int           `mapstructure:"max_login_attempts"`
	LockoutDuration      time.Duration `mapstructure:"lockout_duration"`
}

type AWSConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	SessionToken    string `mapstructure:"session_token"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/anybase")
	}

	// Set defaults
	setDefaults()

	// Bind environment variables
	viper.SetEnvPrefix("ANYBASE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Use defaults if config file not found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	cfg = &config
	return cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "development")
	viper.SetDefault("server.read_timeout", 15*time.Second)
	viper.SetDefault("server.write_timeout", 15*time.Second)
	viper.SetDefault("server.shutdown_timeout", 10*time.Second)

	// Database defaults
	viper.SetDefault("database.uri", "mongodb://localhost:27017")
	viper.SetDefault("database.database", "anybase")
	viper.SetDefault("database.max_pool_size", 100)
	viper.SetDefault("database.min_pool_size", 10)
	viper.SetDefault("database.max_idle_time", 10*time.Minute)
	viper.SetDefault("database.heartbeat_interval", 10*time.Second)
	viper.SetDefault("database.retry_writes", true)
	viper.SetDefault("database.server_selection_timeout", 5*time.Second)

	// Auth defaults
	viper.SetDefault("auth.jwt_expiration", 24*time.Hour)
	viper.SetDefault("auth.refresh_token_expiration", 7*24*time.Hour)
	viper.SetDefault("auth.password_min_length", 8)
	viper.SetDefault("auth.bcrypt_cost", 10)
	viper.SetDefault("auth.max_login_attempts", 5)
	viper.SetDefault("auth.lockout_duration", 15*time.Minute)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output_path", "stdout")
}

func Get() *Config {
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg
}