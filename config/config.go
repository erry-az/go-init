package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL string `mapstructure:"database_url"`
}

// New loads the config file into Config struct
func New() (*Config, error) {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// More option of config path can be added here
	viper.AddConfigPath("/app/config/") // Staging, Production or Docker
	viper.AddConfigPath("files/")       // Unix Local
	viper.AddConfigPath("../../files/") // Windows Local

	viper.AutomaticEnv()

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
