// Package nacos 封装 Nacos 配置中心客户端。
//
// 本项目将 Nacos 作为「强依赖」的配置来源：启动时必须能连上 Nacos 并成功拉取到
// 指定 DataId 的配置，否则直接启动失败（禁止以缺失配置的状态运行）。
package nacos

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
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

// ValidateConnection 校验连接 Nacos 服务端（注册中心/配置中心共用）所需的
// 最小参数：服务地址与端口。命名空间、分组、DataId 等按需由各能力自行校验。
func (c Config) ValidateConnection() error {
	var miss []string
	if c.IpAddr == "" {
		miss = append(miss, "NACOS_IP")
	}
	if c.Port == 0 {
		miss = append(miss, "NACOS_PORT")
	}
	if len(miss) > 0 {
		return fmt.Errorf("nacos connection params are missing: %s", strings.Join(miss, ", "))
	}
	return nil
}

// HasConnectionParams 判断 env 中是否提供了 Nacos 连接参数（NACOS_IP/PORT）。
// 这是「是否配置 Nacos」的必要前提；最终是否生效还需 Nacos 能正常连接。
func (c Config) HasConnectionParams() bool {
	return c.ValidateConnection() == nil
}

// ValidateConfigCenter 在连接参数基础上，额外校验「配置中心」所需的分组与 DataId。
// 仅当启用统一配置中心（即设置了 NACOS_DATA_ID）时才需要满足。
func (c Config) ValidateConfigCenter() error {
	if err := c.ValidateConnection(); err != nil {
		return err
	}
	var miss []string
	if c.Group == "" {
		miss = append(miss, "NACOS_GROUP")
	}
	if c.DataID == "" {
		miss = append(miss, "NACOS_DATA_ID")
	}
	if len(miss) > 0 {
		return fmt.Errorf("nacos config-center params are missing: %s", strings.Join(miss, ", "))
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
	if err := cfg.ValidateConfigCenter(); err != nil {
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
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
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

// NamingClient 包装 Nacos 注册中心客户端，用于服务发现（动态获取下游实例地址）。
type NamingClient struct {
	inner  naming_client.INamingClient
	config Config
}

// NewNamingClient 基于同一份 Nacos 连接参数创建注册中心客户端。
// 注意：服务发现与配置中心共用 Nacos 地址/命名空间，但属于不同客户端实例。
func NewNamingClient(cfg Config) (*NamingClient, error) {
	if err := cfg.ValidateConnection(); err != nil {
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

	inner, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos naming client: %w", err)
	}
	return &NamingClient{inner: inner, config: cfg}, nil
}

// Instance 描述一个可用的下游服务实例。
type Instance struct {
	IP       string
	Port     uint64
	Metadata map[string]string
}

// Addr 返回可被直接拼接的 host:port 地址。
func (i Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.IP, i.Port)
}

// RegisterInstance 把本服务实例注册到 Nacos 注册中心。
// serviceName 一般为本服务名（如 dextea-customer-api）；group 为空时使用默认分组。
// SDK 注册成功后会按配置自动发送心跳维持健康状态，无需业务层手动保活。
func (c *NamingClient) RegisterInstance(serviceName, group, ip string, port uint64, metadata map[string]string) error {
	if group == "" {
		group = c.config.Group
	}
	_, err := c.inner.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: serviceName,
		GroupName:   group,
		Ip:          ip,
		Port:        port,
		Weight:      1,
		Healthy:     true,
		Enable:      true,
		Ephemeral:   true, // 临时实例：进程退出后 Nacos 自动剔除，无需强依赖注销
		Metadata:    metadata,
	})
	if err != nil {
		return fmt.Errorf("register instance (service=%s, group=%s, addr=%s:%d): %w", serviceName, group, ip, port, err)
	}
	return nil
}

// DeregisterInstance 从 Nacos 注册中心注销本服务实例（优雅关停时调用）。
// 临时实例即使不注销也会超时剔除，但主动注销可让其它消费者更快感知下线。
func (c *NamingClient) DeregisterInstance(serviceName, group, ip string, port uint64) error {
	if group == "" {
		group = c.config.Group
	}
	_, err := c.inner.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: serviceName,
		GroupName:   group,
		Ip:          ip,
		Port:        port,
	})
	if err != nil {
		return fmt.Errorf("deregister instance (service=%s, group=%s, addr=%s:%d): %w", serviceName, group, ip, port, err)
	}
	return nil
}

// SelectInstance 通过 Nacos 服务发现选取一个健康实例。
// serviceName 为注册在 Nacos 上的服务名；group 为空时回退到配置中的默认分组。
func (c *NamingClient) SelectInstance(serviceName, group string) (*Instance, error) {
	if group == "" {
		group = c.config.Group
	}
	instances, err := c.inner.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   group,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos select instances (service=%s, group=%s): %w", serviceName, group, err)
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("no healthy instance found for service=%s, group=%s", serviceName, group)
	}
	// 在健康实例中随机挑选，实现简单的客户端负载均衡。
	idx := randInt(len(instances))
	ins := instances[idx]
	return &Instance{
		IP:       ins.Ip,
		Port:     ins.Port,
		Metadata: ins.Metadata,
	}, nil
}

// Resolver 把「服务名 → 实时地址」的解析逻辑封装起来，供业务层在每次请求时动态寻址。
// 内部带短期本地缓存，避免对 Nacos 的高频请求；一旦解析失败则用上一次成功结果兜底。
type Resolver struct {
	client *NamingClient
	name   string
	group  string

	mu       sync.RWMutex
	cached   string
	lastSync time.Time
}

// NewResolver 创建一个服务名解析器。
func NewResolver(client *NamingClient, serviceName, group string) *Resolver {
	return &Resolver{client: client, name: serviceName, group: group}
}

// Resolve 返回当前可用的实例地址（host:port）。
// 解析失败且没有缓存时返回 error；有缓存时返回最后一次成功结果，保证可用性。
func (r *Resolver) Resolve() (string, error) {
	r.mu.RLock()
	cached, lastSync := r.cached, r.lastSync
	r.mu.RUnlock()

	// 缓存有效期 3s：短 TTL 保证地址变更能被较快感知，又不过度打扰 Nacos。
	if cached != "" && time.Since(lastSync) < 3*time.Second {
		return cached, nil
	}

	ins, err := r.client.SelectInstance(r.name, r.group)
	if err != nil {
		if cached != "" {
			// 解析失败但有缓存，降级返回旧地址，避免请求整体失败。
			return cached, nil
		}
		return "", err
	}

	addr := ins.Addr()
	r.mu.Lock()
	r.cached = addr
	r.lastSync = time.Now()
	r.mu.Unlock()
	return addr, nil
}

var randSeed uint64

// randInt 返回一个 [0, n) 的伪随机数，用原子自增 + 取模实现无锁随机挑选。
func randInt(n int) int {
	if n <= 1 {
		return 0
	}
	v := atomic.AddUint64(&randSeed, 1)
	return int(v % uint64(n))
}

// Registrar 负责把本服务自身注册到 Nacos 注册中心，并在关停时注销。
// 注册为「软依赖」：Nacos 不可用时注册失败仅告警，不阻断服务启动。
type Registrar struct {
	client   *NamingClient
	name     string
	ip       string
	port     uint64
	metadata map[string]string
}

// NewRegistrar 构造本服务注册器。
//
// 「是否配置 Nacos」由 env 中的连接参数决定：
//   - env 未提供 NACOS_IP/PORT（或 name 为空）→ 视为未配置，返回 (nil, nil)。
//   - env 提供了连接参数 → 构造 naming 客户端，准备注册；随后 Register() 会真正
//     连接 Nacos，只有连接成功才算「配了 Nacos」并完成注册。连接失败（Nacos 不可达）
//     等同未配置，由调用方按「回退 .env、不注册」处理。
func NewRegistrar(cfg Config, name string, port uint64, metadata map[string]string) (*Registrar, error) {
	if name == "" {
		return nil, nil
	}
	if !cfg.HasConnectionParams() {
		// env 未配置 Nacos 连接参数：视为未配置，不注册。
		return nil, nil
	}
	client, err := NewNamingClient(cfg)
	if err != nil {
		return nil, err
	}
	ip := outboundIP()
	return &Registrar{
		client:   client,
		name:     name,
		ip:       ip,
		port:     port,
		metadata: metadata,
	}, nil
}

// Addr 返回本实例注册用的 host:port。
func (r *Registrar) Addr() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", r.ip, r.port)
}

// Register 执行注册。错误由调用方按软依赖处理（告警即可）。
func (r *Registrar) Register() error {
	if r == nil {
		return nil
	}
	return r.client.RegisterInstance(r.name, r.client.config.Group, r.ip, r.port, r.metadata)
}

// Deregister 执行注销。
func (r *Registrar) Deregister() error {
	if r == nil {
		return nil
	}
	return r.client.DeregisterInstance(r.name, r.client.config.Group, r.ip, r.port)
}

// outboundIP 获取本机对外的非回环 IPv4 地址，用于注册到 Nacos 供其它服务访问。
// 取不到时回退到 127.0.0.1。
func outboundIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			return ip4.String()
		}
	}
	return "127.0.0.1"
}
