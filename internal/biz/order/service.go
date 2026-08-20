package order

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/pkg/consts"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const orderHTTPTimeout = 10 * time.Second

// ForwardResult 封装下游订单服务返回的原始响应，由 handler 透传写出。
type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Service 负责把 Order 模块中已与下游一一对应的接口请求转发到 Java 订单服务。
// 自身不执行业务逻辑，仅完成统一的身份透传（从 X-Customer-Id 头）与请求转发。
type Service struct {
	httpClient *http.Client
	baseURL    string
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		httpClient: &http.Client{Timeout: orderHTTPTimeout},
		baseURL:    strings.TrimRight(cfg.OrderServiceBaseURL, "/"),
	}
}

// Forward 透传单个已与下游一一对应的接口请求到 Java 订单服务。
//
// path 为原始请求路径（含 Go 服务自身前缀，如 /api/v1/orders/123），转发时会
// 剥离 /api/v1 前缀后再拼接 baseURL；method、rawQuery、body 原样透传。
// 顾客身份通过固定头 X-Customer-Id 透传（已由全局 Auth 中间件解析并写入请求头），
// 供下游据此做数据归属校验，不再从请求体覆盖顾客标识字段。
func (s *Service) Forward(ctx context.Context, customerID int64, method, path, rawQuery string, body []byte) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(ErrOrderServiceNotConfigured)
	}

	target := s.baseURL + stripAPIPrefix(path)
	if rawQuery != "" {
		target += "?" + rawQuery
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

	return doForward(s.httpClient, req)
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

// doForward 执行下游请求并原样回传响应体，集中处理网络异常。
func doForward(client *http.Client, req *http.Request) (*ForwardResult, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, bizerror.New(ErrOrderServiceUnavailable, "请求订单服务失败")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerror.New(ErrOrderServiceUnavailable, "读取订单服务响应失败")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json; charset=utf-8"
	}

	return &ForwardResult{
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Body:        respBody,
	}, nil
}
