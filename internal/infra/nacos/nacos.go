// Package nacos 封装 Nacos 配置中心客户端。
//
// 本项目将 Nacos 作为「强依赖」的配置来源：启动时必须能连上 Nacos 并成功拉取到
// 指定 DataId 的配置，否则直接启动失败（禁止以缺失配置的状态运行）。
package nacos

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Config 描述连接 Nacos 所需的参数，均由环境变量注入。
type Config struct {
	IpAddr      string // Nacos 服务地址（必填）
	Port        uint64 // Nacos 服务端口（必填）
	NamespaceID string // 命名空间 ID，public 命名空间填空字符串
	Group       string // 配置分组（必填）
	DataID      string // 配置 DataId（必填）
	Username    string // 服务端 API 鉴权用户名（可选）
	Password    string // 服务端 API 鉴权密码（可选）
	TimeoutMs   uint64 // 请求超时，默认 5000ms
	LogLevel    string // 日志级别 debug/info/warn/error，默认 info
}

// Validate 校验 Nacos 连接所需的核心参数是否齐备。
func (c Config) Validate() error {
	var miss []string
	if c.IpAddr == "" {
		miss = append(miss, "NACOS_IP")
	}
	if c.Port == 0 {
		miss = append(miss, "NACOS_PORT")
	}
	if c.Group == "" {
		miss = append(miss, "NACOS_GROUP")
	}
	if c.DataID == "" {
		miss = append(miss, "NACOS_DATA_ID")
	}
	if len(miss) > 0 {
		return fmt.Errorf("nacos is required but connection params are missing: %s", strings.Join(miss, ", "))
	}
	return nil
}

// Client 包装 Nacos 配置客户端。
type Client struct {
	inner  config_client.IConfigClient
	config Config
}

// NewClient 创建并校验 Nacos 配置客户端。
//
// 仅完成客户端构造；是否可达由 Load() 中的 GetConfig 实际请求保证。
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = 5000
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.NamespaceID),
		constant.WithTimeoutMs(cfg.TimeoutMs),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel(cfg.LogLevel),
		constant.WithUsername(cfg.Username),
		constant.WithPassword(cfg.Password),
	)
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.IpAddr, cfg.Port, constant.WithScheme("http")),
	}

	inner, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos config client: %w", err)
	}
	return &Client{inner: inner, config: cfg}, nil
}

// Load 从 Nacos 拉取配置内容并解析为 KEY=VALUE 映射（dotenv 风格）。
//
// 这是「强依赖」的核心：连接失败或 DataId 不存在/为空都会返回 error，
// 调用方应据此中止启动。
func (c *Client) Load() (map[string]string, error) {
	content, err := c.inner.GetConfig(vo.ConfigParam{
		DataId: c.config.DataID,
		Group:  c.config.Group,
	})
	if err != nil {
		return nil, fmt.Errorf("get nacos config (dataId=%s, group=%s): %w", c.config.DataID, c.config.Group, err)
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("nacos config is empty (dataId=%s, group=%s)", c.config.DataID, c.config.Group)
	}
	return parseDotenv(content), nil
}

// parseDotenv 将 dotenv 风格文本解析为键值映射，忽略空行与注释（# 开头）。
func parseDotenv(content string) map[string]string {
	out := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}
		out[key] = val
	}
	return out
}

// LoadFromEnv 从环境变量读取 Nacos 连接配置。
func LoadFromEnv() Config {
	return Config{
		IpAddr:      os.Getenv("NACOS_IP"),
		Port:       uint64(atoiDefault(os.Getenv("NACOS_PORT"), 0)),
		NamespaceID: os.Getenv("NACOS_NAMESPACE_ID"),
		Group:       getenvDefault("NACOS_GROUP", "DEFAULT_GROUP"),
		DataID:      os.Getenv("NACOS_DATA_ID"),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
		TimeoutMs:   uint64(atoiDefault(os.Getenv("NACOS_TIMEOUT_MS"), 5000)),
		LogLevel:    getenvDefault("NACOS_LOG_LEVEL", "info"),
	}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
