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
		z, _ := zapCfg.Build()
		logger = z.Sugar()
	})
	return logger
}
