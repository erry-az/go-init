package config

type ServerConfig struct {
	GrpcPort string `mapstructure:"grpc_port" default:"8080"`
	HttpPort string `mapstructure:"http_port" default:"8081"`
}
