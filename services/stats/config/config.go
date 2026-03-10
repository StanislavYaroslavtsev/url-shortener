package config

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
}

type ClickHouseConfig struct {
	Addr     string `mapstructure:"addr"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type KafkaConfig struct {
	Addr    string `mapstructure:"addr"`
	Topic   string `mapstructure:"topic"`
	GroupID string `mapstructure:"group_id"`
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
			"clickhouse_addr", cfg.ClickHouse.Addr,
			"kafka_addr", cfg.Kafka.Addr,
			"kafka_topic", cfg.Kafka.Topic,
		)
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func setDefaults() {
	viper.SetDefault("clickhouse.addr", "clickhouse:9000")
	viper.SetDefault("clickhouse.database", "stats")
	viper.SetDefault("clickhouse.user", "stats")
	viper.SetDefault("clickhouse.password", "stats")

	viper.SetDefault("kafka.addr", "kafka:29092")
	viper.SetDefault("kafka.topic", "click_events")
	viper.SetDefault("kafka.group_id", "stats")
}

func bindEnvVars() {
	_ = viper.BindEnv("clickhouse.addr", "CLICKHOUSE_ADDR")
	_ = viper.BindEnv("clickhouse.database", "CLICKHOUSE_DATABASE")
	_ = viper.BindEnv("clickhouse.user", "CLICKHOUSE_USER")
	_ = viper.BindEnv("clickhouse.password", "CLICKHOUSE_PASSWORD")

	_ = viper.BindEnv("kafka.addr", "KAFKA_ADDR")
	_ = viper.BindEnv("kafka.topic", "KAFKA_TOPIC")
	_ = viper.BindEnv("kafka.group_id", "KAFKA_GROUP_ID")
}
