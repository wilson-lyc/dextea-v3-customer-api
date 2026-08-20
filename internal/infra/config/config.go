package config

import (
	"fmt"
	"log"
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

	// 订单服务在 Nacos 注册中心注册的服务名（用于动态服务发现）。
	// 配置后，调用订单服务时实时从 Nacos 拉取健康实例地址，而非写死地址。
	// 与 OrderServiceBaseURL 同时存在时，以服务发现优先；两者皆缺则启动自检报错。
	OrderServiceName string
	// 订单服务在 Nacos 注册中心所属分组，为空时使用默认分组（同 Nacos 配置分组）。
	OrderServiceGroup string

	// nacosRaw 保存 Nacos 连接参数，供业务层（如订单服务发现）构造 naming 客户端。
	nacosRaw nacos.Config
}

// Load 加载配置。
//
// 是否为「Nacos 模式」由是否配置 Nacos 服务端地址(NACOS_IP/PORT) 决定：
//   - 未配置 Nacos：既不在 Nacos 注册本服务，也不读取 Nacos 配置，纯 .env / 默认值运行。
//   - 已配置 Nacos：本服务会注册到 Nacos（见 main）；并若同时配置了 NACOS_DATA_ID，
//     则以「本地环境变量(.env) > Nacos 配置 > 默认值」的优先级合并配置（env 优先）。
//
// Nacos 配置中心不可达或拉取失败时不再阻塞启动（软依赖，仅告警），回退到 .env / 默认值。
func Load() (*Config, error) {
	_ = godotenv.Load()

	nacosValues, err := loadNacosConfig()
	if err != nil {
		// 配置中心为软依赖：拉取失败仅告警，不中断启动。
		log.Printf("[WARN] nacos config-center unavailable, fall back to .env / defaults: %v", err)
		nacosValues = nil
	}

	l := newLookup(nacosValues)

	return &Config{
		nacosRaw: nacos.LoadFromEnv(),
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
		OrderServiceName:    l.get("ORDER_SERVICE_NAME", ""),
		OrderServiceGroup:   l.get("ORDER_SERVICE_GROUP", ""),
	}, nil
}

// loadNacosConfig 建立 Nacos 配置中心客户端并拉取配置，返回 dotenv 风格的键值映射。
//
// 启用条件（与「本服务是否注册到 Nacos」共用同一开关）：必须同时配置了 Nacos 服务
// 端地址(NACOS_IP/PORT) 与 配置项(NACOS_DATA_ID)。也即「没有配置 Nacos 就不读 Nacos
// 配置」，此时返回 (nil, nil)，调用方完全依赖 .env 与默认值。
// 启用后若客户端创建或拉取失败，返回 error，由 Load() 以软依赖方式降级处理（不阻断启动）。
func loadNacosConfig() (map[string]string, error) {
	nacosCfg := nacos.LoadFromEnv()
	// 未配置 Nacos 服务端地址（env 中无 NACOS_IP/PORT）或未指定 DataId：
	// 视为未配置 Nacos，不读取其配置，直接跳过。
	if !nacosCfg.HasConnectionParams() || nacosCfg.DataID == "" {
		return nil, nil
	}
	client, err := nacos.NewClient(nacosCfg)
	if err != nil {
		return nil, fmt.Errorf("nacos config-center init: %w", err)
	}
	values, err := client.Load()
	if err != nil {
		return nil, fmt.Errorf("nacos config-center load: %w", err)
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

	if c.OrderServiceName == "" && c.OrderServiceBaseURL == "" {
		hasErr = true
		b.WriteString("\n  [Order] order service target is not configured, please add ONE of the following in .env or environment variables:")
		b.WriteString("\n    - ORDER_SERVICE_NAME (preferred, dynamic discovery via Nacos registry)")
		b.WriteString("\n    - ORDER_SERVICE_BASE_URL (fallback, static base URL)")
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

// NacosConfig 返回 Nacos 连接参数，供需要注册中心能力的业务层复用同一连接配置。
func (c *Config) NacosConfig() nacos.Config {
	return c.nacosRaw
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
