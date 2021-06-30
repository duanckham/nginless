package main

import (
	"github.com/duanckham/nginless/internal/app/nginless"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// Log rotate.
	// Refs:
	// https://github.com/uber-go/zap/blob/master/FAQ.md#does-zap-support-log-rotation
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "/tmp/nginless.log",
		MaxSize:    1000,
		MaxBackups: 20,
		MaxAge:     14,
	})

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zap.InfoLevel,
	))

	defer logger.Sync()

	n := nginless.New(nginless.Options{
		Logger: logger,
	})

	n.Run()
}
