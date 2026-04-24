package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	OllamaURL          string `mapstructure:"ollama_url"`
	Model              string `mapstructure:"model"`
	Namespace          string `mapstructure:"namespace"`
	LogLevel           string `mapstructure:"log_level"`
	MaxLogLines        int    `mapstructure:"max_log_lines"`
	RequestTimeout     int    `mapstructure:"request_timeout"`
	OllamaRetries      int    `mapstructure:"ollama_retries"`
	OllamaBackoffMs    int    `mapstructure:"ollama_backoff_ms"`
	OllamaMaxBackoffMs int    `mapstructure:"ollama_max_backoff_ms"`
}

var cfg *Config

func GetConfig() *Config {
	if cfg == nil {
		cfg = loadConfig()
	}
	return cfg
}

func loadConfig() *Config {
	viper.SetDefault("ollama_url", "http://127.0.0.1:11434")
	viper.SetDefault("model", "gemma4:e4b")
	viper.SetDefault("namespace", "default")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("max_log_lines", 100)
	viper.SetDefault("request_timeout", 60)
	viper.SetDefault("ollama_retries", 3)
	viper.SetDefault("ollama_backoff_ms", 500)
	viper.SetDefault("ollama_max_backoff_ms", 5000)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".config", "clusterbrain"))
	}
	viper.AddConfigPath(".")

	// Ignore error if config file doesn't exist
	_ = viper.ReadInConfig()

	var c Config
	_ = viper.Unmarshal(&c)
	return &c
}
