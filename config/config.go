package config

import (
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	HandlerTimeout  time.Duration `mapstructure:"handler_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type AppConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Env     string `mapstructure:"env"`
}

type DatabaseConfig struct {
	UsePostgres bool   `mapstructure:"use_postgres"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	SSLMode     string `mapstructure:"ssl_mode"`
}

type CacheConfig struct {
	TTL             time.Duration `mapstructure:"ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

var (
	instance *Config
	once     sync.Once
)

func Init() (*Config, error) {
	var initErr error

	once.Do(func() {

		setDefaults()

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		bindEnvVars()

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
				slog.Warn("Config file not found, using defaults")
			} else {
				initErr = err
				return
			}
		}

		cfg := &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			initErr = err
			return
		}

		instance = cfg

		slog.Info("Config loaded",
			"env", cfg.App.Env,
			"use_postgres", cfg.Database.UsePostgres,
			"server_port", cfg.Server.Port,
		)
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.handler_timeout", "30s")
	viper.SetDefault("server.shutdown_timeout", "5s")

	viper.SetDefault("app.base_url", "http://localhost:3000")
	viper.SetDefault("app.env", "development")

	viper.SetDefault("database.use_postgres", false)
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "url_shortener")
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("cache.ttl", "24h")
	viper.SetDefault("cache.cleanup_interval", "1m")
}

func bindEnvVars() {
	_ = viper.BindEnv("database.host", "POSTGRES_HOST")
	_ = viper.BindEnv("database.port", "POSTGRES_PORT")
	_ = viper.BindEnv("database.user", "POSTGRES_USER")
	_ = viper.BindEnv("database.password", "POSTGRES_PASSWORD")
	_ = viper.BindEnv("database.dbname", "POSTGRES_DB")
	_ = viper.BindEnv("database.ssl_mode", "POSTGRES_SSL_MODE")
	_ = viper.BindEnv("app.env", "APP_ENV")
}
