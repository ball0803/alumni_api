package logger

import (
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	config := zap.Config{
		Encoding:         "console", // Switch to "console" for more readable output
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
	}
	return config.Build()
}
