package config

type DatabaseConfig struct {
	DbDsn   string `mapstructure:"db_dsn"`
	PgMqUrl string `mapstructure:"pg_mq"`
}
