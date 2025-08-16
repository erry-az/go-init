package config

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

const (
	RetryConsumerTypeDefault      = "default"
	RetryConsumerTypeConservative = "conservative"
	RetryConsumerTypeAggressive   = "aggressive"
)

type ConsumerConfig struct {
	Retry *RetryConsumerConfig `mapstructure:"retry"`
}

type RetryConsumerConfig struct {
	Type                string        `mapstructure:"type"`
	MaxRetries          int           `mapstructure:"max_retries"`
	InitialInterval     time.Duration `mapstructure:"initial_interval"`
	MaxInterval         time.Duration `mapstructure:"max_interval"`
	Multiplier          float64       `mapstructure:"multiplier"`
	MaxElapsedTime      time.Duration `mapstructure:"max_elapsed_time"`
	RandomizationFactor float64       `mapstructure:"randomization_factor"`
}

// GetRetry replace standard retry behaviour
func (c *RetryConsumerConfig) GetRetry() RetryConsumerConfig {
	ret := DefaultRetryConsumerConfig()
	if c == nil {
		return ret
	}

	switch c.Type {
	case RetryConsumerTypeDefault:
		ret = DefaultRetryConsumerConfig()
	case RetryConsumerTypeConservative:
		ret = ConservativeRetryConsumerConfig()
	case RetryConsumerTypeAggressive:
		ret = AggressiveRetryConsumerConfig()
	}

	if c.MaxRetries > 0 {
		ret.MaxRetries = c.MaxRetries
	}

	if c.InitialInterval > 0 {
		ret.InitialInterval = c.InitialInterval
	}

	if c.MaxInterval > 0 {
		ret.MaxInterval = c.MaxInterval
	}

	if c.Multiplier > 0 {
		ret.Multiplier = c.Multiplier
	}

	if c.MaxElapsedTime > 0 {
		ret.MaxElapsedTime = c.MaxElapsedTime
	}

	if c.RandomizationFactor > 0 {
		ret.RandomizationFactor = c.RandomizationFactor
	}

	return ret
}

func (c *RetryConsumerConfig) MiddlewareRetry(logger watermill.LoggerAdapter) middleware.Retry {
	retrier := c.GetRetry()
	return middleware.Retry{
		MaxRetries:          retrier.MaxRetries,
		InitialInterval:     retrier.InitialInterval,
		MaxInterval:         retrier.MaxInterval,
		Multiplier:          retrier.Multiplier,
		MaxElapsedTime:      retrier.MaxElapsedTime,
		RandomizationFactor: retrier.RandomizationFactor,
		Logger:              logger,
	}
}

func DefaultRetryConsumerConfig() RetryConsumerConfig {
	return RetryConsumerConfig{
		MaxRetries:          3,                      // Reasonable number of retries
		InitialInterval:     100 * time.Millisecond, // Start with 100ms
		MaxInterval:         30 * time.Second,       // Cap at 30 seconds
		Multiplier:          2.0,                    // Double each time (exponential backoff)
		MaxElapsedTime:      5 * time.Minute,        // Give up after 5 minutes total
		RandomizationFactor: 0.1,                    // Add 10% jitter to prevent thundering herd
	}
}

// AggressiveRetryConsumerConfig returns a more aggressive retry configuration
func AggressiveRetryConsumerConfig() RetryConsumerConfig {
	return RetryConsumerConfig{
		MaxRetries:          5,
		InitialInterval:     50 * time.Millisecond,
		MaxInterval:         10 * time.Second,
		Multiplier:          1.5,
		MaxElapsedTime:      2 * time.Minute,
		RandomizationFactor: 0.2,
	}
}

// ConservativeRetryConsumerConfig returns a more conservative retry configuration
func ConservativeRetryConsumerConfig() RetryConsumerConfig {
	return RetryConsumerConfig{
		MaxRetries:          2,
		InitialInterval:     500 * time.Millisecond,
		MaxInterval:         60 * time.Second,
		Multiplier:          3.0,
		MaxElapsedTime:      10 * time.Minute,
		RandomizationFactor: 0.05,
	}
}
