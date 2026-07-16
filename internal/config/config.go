package config

import (
	"os"
	"strconv"
)

// Config 保存应用运行所需的配置项。
type Config struct {
	Port          string
	Environment   string
	ServiceName   string
	DatabaseDriver string
	DatabaseDSN    string
}

// Load 从环境变量中读取配置，未设置时使用默认值。
// DatabaseDSN 默认为空，表示暂不启用数据库。
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		ServiceName:    getEnv("SERVICE_NAME", "dextea-customer-api"),
		DatabaseDriver: getEnv("DATABASE_DRIVER", "sqlite"),
		DatabaseDSN:    getEnv("DATABASE_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
