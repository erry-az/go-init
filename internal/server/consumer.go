package server

import (
	"context"

	"github.com/erry-az/go-sample/internal/app"
)

func NewConsumer(consumerApp *app.ConsumerApp) error {
	return consumerApp.Run(context.Background())
}
