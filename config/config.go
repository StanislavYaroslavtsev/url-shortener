package config

import (
	"errors"
	"log"

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
	} `mapstructure:"app"`
}

func Init() {
	setDefaults()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError

		if errors.As(err, &configFileNotFoundError) {
			log.Println("Config file not found, using defaults")
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	log.Printf("Config loaded")
}

func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 3000)

	viper.SetDefault("app.base_url", "http://localhost:3000")
}

func GetConfig() *Config {
	if AppConfig == nil {
		Init()
	}

	return AppConfig
}
