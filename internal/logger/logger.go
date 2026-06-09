package logger

import (
	"github.com/Luclpor/url_shortener.git/internal/config"
	"go.uber.org/zap"
)

// InitLogger creates a production or development zap logger for the given environment.
func InitLogger(env string) (*zap.Logger, error) {
	switch env {
	case config.ProdEnv:
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}
		return logger, nil
	default:
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		return logger, nil
	}
}
