package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
}

// New loads the config file into Config struct
func New(appName, env string) (Config, error) {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// More option of config path can be added here
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/config/", appName))    // Staging, Production or Docker
	viper.AddConfigPath(fmt.Sprintf("files/config/%s/", env))       // Unix Local
	viper.AddConfigPath(fmt.Sprintf("../../files/config/%s/", env)) // Windows Local

	viper.AutomaticEnv()

	// Get the config file
	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	// Convert into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
