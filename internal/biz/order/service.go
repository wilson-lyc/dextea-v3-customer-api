package order

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/nacos"
	"github.com/dextea-v3/dextea-customer/api/internal/pkg/consts"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const orderHTTPTimeout = 10 * time.Second

// maxRetries 下游瞬时网络失败的有限重试次数（需下游接口幂等）。
const maxRetries = 2

// orderBreaker 保护订单中台调用：连续失败 5 次熔断，冷却 10s 后探测。
var orderBreaker = newCircuitBreaker(5, 10*time.Second)

// ForwardResult 封装下游订单服务返回的原始响应，由 handler 透传写出。
type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Service 负责把 Order 模块中已与下游一一对应的接口请求转发到 Java 订单服务。
// 自身不执行业务逻辑，仅完成统一的身份透传（从 X-Customer-Id 头）与请求转发。
//
// 寻址优先级：配置了 ORDER_SERVICE_NAME 时走 Nacos 注册中心动态服务发现（推荐，
// 地址变更无需重启）；否则回退到 ORDER_SERVICE_BASE_URL 静态地址。
//
// 异常改进（见 docs/异常处理机制重构方案.md）：统一使用 bizerror 错误中心承载错误、
// 携带根因与上下文、对瞬时网络错误做有限重试、并以熔断器保护本服务不被持续不可用的下游拖垮。
type Service struct {
	httpClient *http.Client
	baseURL    string // 静态兜底地址
	resolver   *nacos.Resolver
}

func NewService(cfg *config.Config) *Service {
	svc := &Service{
		httpClient: &http.Client{Timeout: orderHTTPTimeout},
		baseURL:    strings.TrimRight(cfg.OrderServiceBaseURL, "/"),
	}
	// 优先建立 Nacos 服务发现解析器：配置服务名即可，无需直连地址。
	if cfg.OrderServiceName != "" {
		namingClient, err := nacos.NewNamingClient(cfg.NacosConfig())
		if err == nil {
			svc.resolver = nacos.NewResolver(namingClient, cfg.OrderServiceName, cfg.OrderServiceGroup)
		}
	}
	return svc
}

// resolveBaseURL 动态解析订单服务 baseURL。
// 服务发现优先，失败时回退静态地址；两者皆不可用时报错。
func (s *Service) resolveBaseURL() (string, error) {
	if s.resolver != nil {
		if addr, err := s.resolver.Resolve(); err == nil {
			return "http://" + strings.TrimRight(addr, "/"), nil
		}
	}
	if s.baseURL != "" {
		return s.baseURL, nil
	}
	return "", bizerror.New(ErrOrderServiceNotConfigured)
}

// Forward 透传单个已与下游一一对应的接口请求到 Java 订单服务。
//
// path 为原始请求路径（含 Go 服务自身前缀，如 /api/v1/orders/123），转发时会
// 剥离 /api/v1 前缀后再拼接 baseURL；method、rawQuery、body 原样透传。
// 顾客身份通过固定头 X-Customer-Id 透传（已由全局 Auth 中间件解析并写入请求头），
// 供下游据此做数据归属校验，不再从请求体覆盖顾客标识字段。
func (s *Service) Forward(ctx context.Context, customerID int64, method, path, rawQuery string, body []byte) (*ForwardResult, error) {
	baseURL, err := s.resolveBaseURL()
	if err != nil {
		return nil, err
	}

	target := baseURL + stripAPIPrefix(path)
	if rawQuery != "" {
		target += "?" + rawQuery
	}

	if !orderBreaker.allow() {
		return nil, bizerror.New(ErrOrderServiceUnavailable, "下游熔断，暂时拒绝请求")
	}

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, bizerror.New(ErrOrderServiceError, "构建订单服务请求失败")
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	// 透传已认证的顾客身份（固定头 X-Customer-Id），供下游据此做数据归属校验。
	req.Header.Set(consts.CustomerIDHeader, strconv.FormatInt(customerID, 10))

	// 将当前链路上下文（traceparent / trace id）注入请求头，转发给订单中台，
	// 使其能继续同一链路，便于跨服务串联排查问题。
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	result, err := s.doForward(req)
	if err != nil {
		orderBreaker.recordFailure()
		return nil, err
	}
	orderBreaker.recordSuccess()
	return result, nil
}

// doForward 执行下游请求并原样回传响应体。对瞬时网络错误做有限重试（指数退避），
// 但请求构建失败、下游返回非 200 等不可重试错误直接返回。
func (s *Service) doForward(req *http.Request) (*ForwardResult, error) {
	operation := func() (*ForwardResult, error) {
		resp, err := s.httpClient.Do(req)
		if err != nil {
			// 上下文超时/取消不可重试；其余网络瞬时错误可重试。
			if req.Context().Err() != nil {
				return nil, backoff.Permanent(
					bizerror.NewWith(ErrOrderServiceUnavailable, bizerror.WithCause(err)))
			}
			return nil, bizerror.NewWith(ErrOrderServiceUnavailable, bizerror.WithCause(err))
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, backoff.Permanent(
				bizerror.NewWith(ErrOrderServiceError, bizerror.WithCause(err), bizerror.WithMessage("读取订单服务响应失败")))
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/json; charset=utf-8"
		}

		// 下游以 HTTP 状态表达业务结果时，仍透传原响应；仅网络层异常才进入重试/熔断。
		return &ForwardResult{
			StatusCode:  resp.StatusCode,
			ContentType: contentType,
			Body:        respBody,
		}, nil
	}

	result, err := backoff.Retry(req.Context(), operation,
		backoff.WithMaxTries(maxRetries+1),
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
	)
	if err != nil {
		if _, ok := bizerror.As(err); ok {
			return nil, err
		}
		return nil, bizerror.NewWith(ErrOrderServiceUnavailable, bizerror.WithCause(err))
	}
	return result, nil
}

// stripAPIPrefix 剥离请求路径中 Go 服务自身的 API 前缀，避免与下游 baseURL 重复。
// 仅当路径等于该前缀或其后的下一段紧跟 "/" 时才剥离（如 /api/v1/orders），
// 避免误伤 /api/v10 这类以相同前缀开头的路径；剥离结果为空时回退为 "/"。
func stripAPIPrefix(path string) string {
	p := path
	switch {
	case p == consts.APIPrefixV1:
		return "/"
	case strings.HasPrefix(p, consts.APIPrefixV1+"/"):
		p = strings.TrimPrefix(p, consts.APIPrefixV1)
	}
	if p == "" {
		return "/"
	}
	return p
}
