package logger

import (
	"go.uber.org/zap"
	"os"
)

func NewLogger() (*zap.Logger, error) {
	var config zap.Config

	if os.Getenv("ENV") == "dev" {
		config = zap.Config{
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		}
	} else {
		config = zap.Config{
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		}
	}

	return config.Build()
}
