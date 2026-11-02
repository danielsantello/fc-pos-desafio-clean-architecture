package configs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type conf struct {
	DBDriver          string `mapstructure:"DB_DRIVER"`
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	WebServerPort     string `mapstructure:"WEB_SERVER_PORT"`
	GRPCServerPort    string `mapstructure:"GRPC_SERVER_PORT"`
	GraphQLServerPort string `mapstructure:"GRAPHQL_SERVER_PORT"`
	RabbitMQURL       string `mapstructure:"RABBITMQ_URL"`
}

func LoadConfig(path string) (*conf, error) {
	config := viper.New()
	config.SetConfigFile(filepath.Join(path, ".env"))
	config.SetConfigType("env")

	keys := []string{
		"DB_DRIVER",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"WEB_SERVER_PORT",
		"GRPC_SERVER_PORT",
		"GRAPHQL_SERVER_PORT",
		"RABBITMQ_URL",
	}

	for _, key := range keys {
		if err := config.BindEnv(key); err != nil {
			return nil, fmt.Errorf("bind environment variable %s: %w", key, err)
		}
	}

	config.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	if err := config.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read configuration: %w", err)
		}
	}

	var cfg conf
	if err := config.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}

	return &cfg, nil
}
