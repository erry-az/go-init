package watermill

import "time"

type Config struct {
	AMQPURL      string
	Exchange     string
	ExchangeType string
	Durable      bool
	QueueConfig  QueueConfig
}

type QueueConfig struct {
	GenerateName func(topic string) string
	Durable      bool
	AutoDelete   bool
	Exclusive    bool
	NoWait       bool
	Arguments    map[string]interface{}
	MaxLength    int
	TTL          time.Duration
}

func DefaultConfig(amqpURL string) *Config {
	return &Config{
		AMQPURL:      amqpURL,
		Exchange:     "events",
		ExchangeType: "topic",
		Durable:      true,
		QueueConfig: QueueConfig{
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
		},
	}
}