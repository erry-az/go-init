package config

type ServerConfig struct {
	GrpcPort string `envconfig:"GRPC_PORT" default:"8080"`
	HttpPort string `envconfig:"HTTP_PORT" default:"8081"`
}
