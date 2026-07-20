package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
	ServiceName string

	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBCharset        string
	DBParseTime      bool
	DBLoc            string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	AlipayAppID      string
	AlipayPrivateKey string
	AlipayPublicKey  string
	AlipayGateway    string

	AmapAPIKey string

	JWTSecret      string
	JWTExpireHours int

	// 订单模块自身不执行业务，仅将请求转发到 Java 订单服务。
	// OrderServiceBaseURL 为 Java 订单服务基础地址，各接口的转发路径独立配置。
	OrderServiceBaseURL string
	OrderCalculatePath  string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: getEnv("SERVICE_NAME", "dextea-customer-api"),

		DBHost:         getEnv("DB_HOST", ""),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", ""),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", ""),
		DBCharset:      getEnv("DB_CHARSET", "utf8mb4"),
		DBParseTime:    getEnvBool("DB_PARSE_TIME", true),
		DBLoc:          getEnv("DB_LOC", "Local"),

		RedisAddr:     getEnv("REDIS_ADDR", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		AlipayAppID:      getEnv("ALIPAY_APP_ID", ""),
		AlipayPrivateKey: getEnv("ALIPAY_PRIVATE_KEY", ""),
		AlipayPublicKey:  getEnv("ALIPAY_PUBLIC_KEY", ""),
		AlipayGateway:    getEnv("ALIPAY_GATEWAY", ""),

		AmapAPIKey: getEnv("AMAP_API_KEY", ""),

		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 168),

		OrderServiceBaseURL: getEnv("ORDER_SERVICE_BASE_URL", ""),
		OrderCalculatePath:  getEnv("ORDER_CALCULATE_PATH", "/order/calculate"),
	}
}

func (c *Config) Validate() error {
	var b strings.Builder
	hasErr := false

	if c.DBHost == "" || c.DBUser == "" || c.DBName == "" {
		hasErr = true
		b.WriteString("\n  [MySQL] database configuration is incomplete, please add the following in .env or environment variables:")
		if c.DBHost == "" {
			b.WriteString("\n    - DB_HOST")
		}
		if c.DBUser == "" {
			b.WriteString("\n    - DB_USER")
		}
		if c.DBName == "" {
			b.WriteString("\n    - DB_NAME")
		}
	}

	if c.RedisAddr == "" {
		hasErr = true
		b.WriteString("\n  [Redis] cache is not configured, please add the following in .env or environment variables:")
		b.WriteString("\n    - REDIS_ADDR")
	}

	if c.AmapAPIKey == "" {
		hasErr = true
		b.WriteString("\n  [Amap] API Key is not configured, please add the following in .env or environment variables:")
		b.WriteString("\n    - AMAP_API_KEY")
	}

	if c.AlipayAppID == "" || c.AlipayPrivateKey == "" {
		hasErr = true
		b.WriteString("\n  [Alipay] configuration is incomplete, please add the following in .env or environment variables:")
		if c.AlipayAppID == "" {
			b.WriteString("\n    - ALIPAY_APP_ID")
		}
		if c.AlipayPrivateKey == "" {
			b.WriteString("\n    - ALIPAY_PRIVATE_KEY")
		}
	}

	if c.JWTSecret == "" {
		hasErr = true
		b.WriteString("\n  [JWT] secret is not configured, please add the following in .env or environment variables:")
		b.WriteString("\n    - JWT_SECRET")
	}

	if !hasErr {
		return nil
	}

	return fmt.Errorf("startup self-check failed: missing required configuration, startup is forbidden.%s", b.String())
}

func (c *Config) DatabaseDSN() string {
	if c.DBHost == "" || c.DBName == "" {
		return ""
	}

	mc := mysql.NewConfig()
	mc.User = c.DBUser
	mc.Passwd = c.DBPassword
	mc.Net = "tcp"
	mc.Addr = c.DBHost + ":" + c.DBPort
	mc.DBName = c.DBName
	mc.ParseTime = c.DBParseTime

	if c.DBCharset != "" {
		mc.Params = map[string]string{"charset": c.DBCharset}
	}
	if c.DBLoc != "" {
		if loc, err := time.LoadLocation(c.DBLoc); err == nil {
			mc.Loc = loc
		}
	}

	return mc.FormatDSN()
}

func (c *Config) JWTExpireSeconds() int {
	return c.JWTExpireHours * 3600
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

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
