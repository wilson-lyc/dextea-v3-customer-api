package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"github.com/dextea-v3/dextea-customer/api/internal/infra/nacos"
)

type Config struct {
	Port        string
	Environment string
	ServiceName string

	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBCharset   string
	DBParseTime bool
	DBLoc       string

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

	// AuthWhitelist 为免鉴权路径白名单（精确匹配请求路径）
	AuthWhitelist []string

	// 订单模块请求转发路径
	OrderServiceBaseURL string
}

// Load 加载配置。Nacos 为「强依赖」：必须先连上 Nacos 并成功拉取配置，
// 否则返回 error，禁止以缺失配置的状态启动。本地环境变量优先级高于 Nacos 配置，
// 便于在容器/本地环境对个别项做临时覆盖；两者皆缺时使用兜底默认值。
func Load() (*Config, error) {
	_ = godotenv.Load()

	nacosValues, err := loadNacosConfig()
	if err != nil {
		return nil, err
	}

	l := newLookup(nacosValues)

	return &Config{
		Port:        l.get("PORT", "8080"),
		Environment: l.get("ENVIRONMENT", "development"),
		ServiceName: l.get("SERVICE_NAME", "dextea-customer-api"),

		DBHost:      l.get("DB_HOST", ""),
		DBPort:      l.get("DB_PORT", "3306"),
		DBUser:      l.get("DB_USER", ""),
		DBPassword:  l.get("DB_PASSWORD", ""),
		DBName:      l.get("DB_NAME", ""),
		DBCharset:   l.get("DB_CHARSET", "utf8mb4"),
		DBParseTime: l.lookupBool("DB_PARSE_TIME", true),
		DBLoc:       l.get("DB_LOC", "Local"),

		RedisAddr:     l.get("REDIS_ADDR", ""),
		RedisPassword: l.get("REDIS_PASSWORD", ""),
		RedisDB:       l.lookupInt("REDIS_DB", 0),

		AlipayAppID:      l.get("ALIPAY_APP_ID", ""),
		AlipayPrivateKey: l.get("ALIPAY_PRIVATE_KEY", ""),
		AlipayPublicKey:  l.get("ALIPAY_PUBLIC_KEY", ""),
		AlipayGateway:    l.get("ALIPAY_GATEWAY", ""),

		AmapAPIKey: l.get("AMAP_API_KEY", ""),

		JWTSecret:      l.get("JWT_SECRET", ""),
		JWTExpireHours: l.lookupInt("JWT_EXPIRE_HOURS", 168),

		AuthWhitelist: l.lookupList("AUTH_WHITELIST", []string{"/api/v1/customers/login"}),

		OrderServiceBaseURL: l.get("ORDER_SERVICE_BASE_URL", ""),
	}, nil
}

// loadNacosConfig 建立 Nacos 客户端并拉取配置，返回 dotenv 风格的键值映射。
// Nacos 不可达或 DataId 缺失/为空均返回 error（强依赖）。
func loadNacosConfig() (map[string]string, error) {
	nacosCfg := nacos.LoadFromEnv()
	client, err := nacos.NewClient(nacosCfg)
	if err != nil {
		return nil, fmt.Errorf("nacos required: %w", err)
	}
	values, err := client.Load()
	if err != nil {
		return nil, fmt.Errorf("nacos required: %w", err)
	}
	return values, nil
}

// lookup 合并「本地环境变量 > Nacos 配置 > 默认值」的只读取值器。
type lookup struct {
	nacos map[string]string
}

func newLookup(nacosValues map[string]string) *lookup {
	return &lookup{nacos: nacosValues}
}

func (l *lookup) get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	if v, ok := l.nacos[key]; ok {
		return v
	}
	return fallback
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

func (l *lookup) lookupBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	if v := l.nacos[key]; v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func (l *lookup) lookupInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	if v := l.nacos[key]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func (l *lookup) lookupList(key string, fallback []string) []string {
	v := l.get(key, "")
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
