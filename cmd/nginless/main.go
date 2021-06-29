package main

import (
	"github.com/duanckham/nginless/internal/app/nginless"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	n := nginless.New(nginless.Options{
		Logger: logger,
	})

	n.Run()
}
