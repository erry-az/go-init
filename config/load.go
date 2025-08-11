package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL string `mapstructure:"database_url"`
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

// Load loads the configuration with default values
func Load() (*Config, error) {
	viper.SetDefault("database_url", "postgres://postgres:postgres@localhost:5432/go_init_db?sslmode=disable")
	
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GO_INIT")
	
	// Bind environment variables
	viper.BindEnv("database_url", "DATABASE_URL")
	
	// Try to read config file (ignore if not found)
	_ = viper.ReadInConfig()
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}
