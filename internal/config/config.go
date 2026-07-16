package config

import (
	"os"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Config 保存应用运行所需的配置项。
type Config struct {
	Port        string
	Environment string
	ServiceName string

	// 数据库配置（MySQL）：拆分为独立字段，DSN 由 DatabaseDSN() 在运行时自动拼接。
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBCharset        string
	DBParseTime      bool
	DBLoc            string // 时区，如 Local / UTC，对应 MySQL DSN 的 loc 参数

	// Redis 配置：RedisAddr 为空表示不启用 Redis。
	RedisAddr     string
	RedisPassword string
	RedisDB       int // 逻辑库编号，默认 0

	// 支付宝登录配置：AppID 与应用私钥必填；公钥用于校验支付宝响应签名（可选）。
	AlipayAppID      string
	AlipayPrivateKey string
	AlipayPublicKey  string
	AlipayGateway    string // 留空使用生产网关 openapi.alipay.com

	// JWT 配置：HS256 签名密钥与令牌有效期。
	JWTSecret      string
	JWTExpireHours int // 令牌有效期（小时），默认 168（7 天）
}

// Load 加载配置：优先读取 .env 文件（若存在），再回退到进程环境变量，最后使用默认值。
// 通过 godotenv 读取项目根目录的 .env，文件不存在时静默忽略（例如生产环境直接注入环境变量）。
func Load() *Config {
	// 忽略文件不存在的错误：在容器/生产环境中通常直接注入环境变量。
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

		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 168),
	}
}

// DatabaseDSN 根据拆分的数据库配置拼接出 MySQL 连接串。
//
// 使用 go-sql-driver/mysql 提供的 mysql.Config 来构建，能自动处理密码等字段中的
// 特殊字符转义，比手写字符串更可靠。生成的 DSN 形如：
//
//	user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=true&loc=Local
//
// 当 DBHost 或 DBName 为空时返回空串，表示不启用数据库，由上层决定降级行为。
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

// JWTExpireSeconds 返回令牌有效期（秒）。
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
