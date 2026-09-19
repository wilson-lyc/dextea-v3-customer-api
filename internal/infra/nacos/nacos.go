package nacos

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
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

type Config struct {
	Enabled     bool
	Host        string
	Port        uint64
	NamespaceID string
	Group       string
	Cluster     string
	DataID      string
	Username    string
	Password    string
	InstanceIP  string
	TimeoutMs   uint64
	LogLevel    string
}

func (c Config) ValidateConnection() error {
	var miss []string
	if c.Host == "" || c.Port == 0 {
		miss = append(miss, "NACOS_SERVER_ADDR")
	}
	if len(miss) > 0 {
		return fmt.Errorf("nacos connection params are missing: %s", strings.Join(miss, ", "))
	}
	return nil
}

func (c Config) HasConnectionParams() bool {
	return c.ValidateConnection() == nil
}

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

type Client struct {
	inner  config_client.IConfigClient
	config Config
}

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
		*constant.NewServerConfig(cfg.Host, cfg.Port, constant.WithScheme("http")),
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

func LoadFromEnv() Config {
	serverAddr := strings.TrimSpace(os.Getenv("NACOS_SERVER_ADDR"))
	host, port := splitServerAddr(serverAddr)
	enabled := serverAddr != ""
	if value, ok := os.LookupEnv("NACOS_ENABLED"); ok && strings.TrimSpace(value) != "" {
		enabled = parseBool(value, enabled)
	}
	return Config{
		Enabled:     enabled,
		Host:        host,
		Port:        port,
		NamespaceID: os.Getenv("NACOS_NAMESPACE"),
		Group:       getenvDefault("NACOS_GROUP", "DEFAULT_GROUP"),
		Cluster:     getenvDefault("NACOS_CLUSTER", "DEFAULT"),
		DataID:      os.Getenv("NACOS_DATA_ID"),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
		InstanceIP:  strings.TrimSpace(os.Getenv("NACOS_INSTANCE_IP")),
	}
}

func splitServerAddr(value string) (string, uint64) {
	if value == "" {
		return "", 0
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return "", 0
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return "", 0
	}
	return host, port
}

func parseBool(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type NamingClient struct {
	inner  naming_client.INamingClient
	config Config
}

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
		*constant.NewServerConfig(cfg.Host, cfg.Port, constant.WithScheme("http")),
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

type Instance struct {
	IP       string
	Port     uint64
	Metadata map[string]string
}

func (i Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.IP, i.Port)
}

func (c *NamingClient) RegisterInstance(serviceName, group, ip string, port uint64, metadata map[string]string) error {
	if group == "" {
		group = c.config.Group
	}
	_, err := c.inner.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: serviceName,
		GroupName:   group,
		ClusterName: c.config.Cluster,
		Ip:          ip,
		Port:        port,
		Weight:      1,
		Healthy:     true,
		Enable:      true,
		Ephemeral:   true,
		Metadata:    metadata,
	})
	if err != nil {
		return fmt.Errorf("register instance (service=%s, group=%s, addr=%s:%d): %w", serviceName, group, ip, port, err)
	}
	return nil
}

func (c *NamingClient) DeregisterInstance(serviceName, group, ip string, port uint64) error {
	if group == "" {
		group = c.config.Group
	}
	_, err := c.inner.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: serviceName,
		GroupName:   group,
		Cluster:     c.config.Cluster,
		Ip:          ip,
		Port:        port,
	})
	if err != nil {
		return fmt.Errorf("deregister instance (service=%s, group=%s, addr=%s:%d): %w", serviceName, group, ip, port, err)
	}
	return nil
}

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
	idx := randInt(len(instances))
	ins := instances[idx]
	return &Instance{
		IP:       ins.Ip,
		Port:     ins.Port,
		Metadata: ins.Metadata,
	}, nil
}

type Resolver struct {
	client *NamingClient
	name   string
	group  string

	mu       sync.RWMutex
	cached   string
	lastSync time.Time
}

func NewResolver(client *NamingClient, serviceName, group string) *Resolver {
	return &Resolver{client: client, name: serviceName, group: group}
}

func (r *Resolver) Resolve() (string, error) {
	r.mu.RLock()
	cached, lastSync := r.cached, r.lastSync
	r.mu.RUnlock()

	if cached != "" && time.Since(lastSync) < 3*time.Second {
		return cached, nil
	}

	ins, err := r.client.SelectInstance(r.name, r.group)
	if err != nil {
		if cached != "" {
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

func randInt(n int) int {
	if n <= 1 {
		return 0
	}
	v := atomic.AddUint64(&randSeed, 1)
	return int(v % uint64(n))
}

type Registrar struct {
	client   *NamingClient
	name     string
	ip       string
	port     uint64
	metadata map[string]string
}

func NewRegistrar(cfg Config, name string, port uint64, metadata map[string]string) (*Registrar, error) {
	if !cfg.Enabled || name == "" {
		return nil, nil
	}
	if !cfg.HasConnectionParams() {
		return nil, nil
	}
	client, err := NewNamingClient(cfg)
	if err != nil {
		return nil, err
	}
	ip := outboundIP()
	if cfg.InstanceIP != "" {
		ip = cfg.InstanceIP
	}
	return &Registrar{
		client:   client,
		name:     name,
		ip:       ip,
		port:     port,
		metadata: metadata,
	}, nil
}

func (r *Registrar) Addr() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", r.ip, r.port)
}

func (r *Registrar) Register() error {
	if r == nil {
		return nil
	}
	return r.client.RegisterInstance(r.name, r.client.config.Group, r.ip, r.port, r.metadata)
}

func (r *Registrar) Deregister() error {
	if r == nil {
		return nil
	}
	return r.client.DeregisterInstance(r.name, r.client.config.Group, r.ip, r.port)
}

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
