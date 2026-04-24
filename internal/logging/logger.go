package logging

import (
	"sync"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		cfg := config.GetConfig()
		level := zapcore.InfoLevel
		switch cfg.LogLevel {
		case "debug":
			level = zapcore.DebugLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		}
		zapCfg := zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(level)
		zapCfg.OutputPaths = []string{"stderr"}
		z, err := zapCfg.Build()
		if err != nil {
			// Fallback to a no-op logger if build fails
			z = zap.NewNop()
		}
		logger = z.Sugar()
	})
	return logger
}
