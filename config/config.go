package config

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

var AppConfig *Config

type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	App struct {
		BaseURL string `mapstructure:"base_url"`
		Env     string `mapstructure:"env"`
	} `mapstructure:"app"`

	Database struct {
		UsePostgres bool   `mapstructure:"use_postgres"`
		Host        string `mapstructure:"host"`
		Port        int    `mapstructure:"port"`
		User        string `mapstructure:"user"`
		Password    string `mapstructure:"password"`
		DBName      string `mapstructure:"dbname"`
		SSLMode     string `mapstructure:"ssl_mode"`
	} `mapstructure:"database"`
}

func Init() {
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
			slog.Error("Failed to read config file", "error", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Error("Unable to decode config",
			"error", err,
		)
	}

	slog.Info("Config loaded",
		"app_env", AppConfig.App.Env,
		"use_postgres", AppConfig.Database.UsePostgres,
	)
}

func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 3000)

	viper.SetDefault("app.base_url", "http://localhost:3000")
	viper.SetDefault("app.env", "development")

	viper.SetDefault("database.use_postgres", false)
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "url_shortener")
	viper.SetDefault("database.ssl_mode", "disable")
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

func GetConfig() *Config {
	if AppConfig == nil {
		Init()
	}

	return AppConfig
}
