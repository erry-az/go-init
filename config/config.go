package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Servers   ServerConfig   `mapstructure:"servers"`
	Databases DatabaseConfig `mapstructure:"databases"`
	Consumers ConsumerConfig `mapstructure:"consumers"`
}

// New loads the config file into Config struct
func New() (*Config, error) {
	var cfg Config

	// Enable environment variable support first
	viper.AutomaticEnv()

	// Check if we're in Docker environment
	configName := "config"
	if isDocker() {
		configName = "config.docker"
	}

	slog.Info("Loading configuration from " + configName)

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// More option of config path can be added here
	viper.AddConfigPath("/app/files/")  // Docker
	viper.AddConfigPath("files/")       // Unix Local
	viper.AddConfigPath("../../files/") // Windows Local

	// Get the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Convert into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// isDocker checks if running in Docker environment
func isDocker() bool {
	// Check common Docker environment indicators
	dockerEnv := viper.GetString("DOCKER_ENV")
	dbHost := viper.GetString("DB_HOST")
	
	slog.Info("Docker environment check", "DOCKER_ENV", dockerEnv, "DB_HOST", dbHost)
	
	if dockerEnv != "" || dbHost != "" {
		return true
	}
	return false
}
