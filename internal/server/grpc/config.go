package grpc

type Config struct {
	Port           string `mapstructure:"port"`
	WithValidation bool   `mapstructure:"with_validation"`
}
