package config

type DatabaseConfig struct {
	GoInitDbUrl string `yaml:"go_init"`
	PgMqUrl     string `yaml:"pg_mq"`
}
