package main

import (
	"github.com/duanckham/nginless/internal/app/config"
	"github.com/duanckham/nginless/internal/app/nginless"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	c := config.ReadConfig()

	// Log rotate.
	// Refs:
	// https://github.com/uber-go/zap/blob/master/FAQ.md#does-zap-support-log-rotation
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   c.Log.Path,
		MaxSize:    c.Log.MaxSize,
		MaxBackups: c.Log.MaxBackups,
		MaxAge:     c.Log.MaxAge,
	})

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zap.InfoLevel,
	))

	defer logger.Sync()

	n := nginless.New(nginless.Options{
		Version: c.Version,
		Logger:  logger,
	})

	n.Run()
}
