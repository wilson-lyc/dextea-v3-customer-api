package order

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

const orderHTTPTimeout = 10 * time.Second

// 订单服务各接口的转发路径（基础地址为 ORDER_SERVICE_BASE_URL，即交易端后台 /api/v1）。
// 这里把 /orders 前缀及各接口路径「写死」在代码中，与交易端 OpenAPI 文档保持一致，
// 不再依赖任何环境变量配置。
const (
	pathOrderList   = "/orders"            // 获取订单列表  GET    /api/v1/orders
	pathOrderCreate = "/orders"            // 创建订单      POST   /api/v1/orders
	pathOrderPreBuild = "/orders/pre-build" // 预构建订单    POST   /api/v1/orders/pre-build
	// 详情 / 状态：/api/v1/orders/{orderId} 与 /api/v1/orders/{orderId}/status
)

// ForwardResult 封装下游订单服务返回的原始响应，由 handler 透传写出。
type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Service 负责把 Order 模块的请求转发到下游 Java 订单服务，自身不执行业务逻辑。
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

// ---- 以下为 5 个路由对应的转发方法 ----

// Create 创建订单：POST /api/v1/orders
func (s *Service) Create(ctx context.Context, customerID int64, body []byte) (*ForwardResult, error) {
	return s.forwardWithBody(ctx, pathOrderCreate, customerID, body)
}

// PreBuild 预构建订单：POST /api/v1/orders/pre-build
func (s *Service) PreBuild(ctx context.Context, customerID int64, body []byte) (*ForwardResult, error) {
	return s.forwardWithBody(ctx, pathOrderPreBuild, customerID, body)
}

// List 获取订单列表：GET /api/v1/orders
func (s *Service) List(ctx context.Context, customerID int64, rawQuery string) (*ForwardResult, error) {
	return s.forwardWithQuery(ctx, pathOrderList, customerID, rawQuery)
}

// Detail 获取订单详情：GET /api/v1/orders/{orderId}
func (s *Service) Detail(ctx context.Context, customerID int64, orderID string, rawQuery string) (*ForwardResult, error) {
	return s.forwardWithQuery(ctx, "/orders/"+orderID, customerID, rawQuery)
}

// Status 获取订单状态：GET /api/v1/orders/{orderId}/status
func (s *Service) Status(ctx context.Context, customerID int64, orderID string, rawQuery string) (*ForwardResult, error) {
	return s.forwardWithQuery(ctx, "/orders/"+orderID+"/status", customerID, rawQuery)
}

// ---- 转发实现 ----

// forwardWithBody 转发带请求体的请求（POST），在写入下游前强制用已认证的
// Customer ID 覆盖请求体中的顾客标识字段，防止越权操作他人数据。
func (s *Service) forwardWithBody(ctx context.Context, path string, customerID int64, body []byte) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(CodeOrderServiceNotConfigured)
	}

	target := s.baseURL + path
	// 用已认证的顾客 id 覆盖请求体中的顾客标识，杜绝 A 的令牌操作 B 的数据。
	body = bindCustomerOwnership(body, customerID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceError, "构建订单服务请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	// 透传已认证的顾客身份（固定头 X-Customer-Id），供下游据此做数据归属校验。
	req.Header.Set(CustomerIDHeader, strconv.FormatInt(customerID, 10))

	return doForward(s.httpClient, req)
}

// forwardWithQuery 转发查询类请求（GET），把原始查询串与路径参数透传，
// 并通过固定头 X-Customer-Id 传递已认证的顾客身份。
func (s *Service) forwardWithQuery(ctx context.Context, path string, customerID int64, rawQuery string) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(CodeOrderServiceNotConfigured)
	}

	target := s.baseURL + path
	if rawQuery != "" {
		target += "?" + rawQuery
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceError, "构建订单服务请求失败")
	}
	// 透传已认证的顾客身份（固定头 X-Customer-Id）。
	req.Header.Set(CustomerIDHeader, strconv.FormatInt(customerID, 10))

	return doForward(s.httpClient, req)
}

// doForward 执行下游请求并原样回传响应体，集中处理网络异常。
func doForward(client *http.Client, req *http.Request) (*ForwardResult, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceUnavailable, "请求订单服务失败")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceUnavailable, "读取订单服务响应失败")
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

// customerIDFields 是请求体中可能出现的「顾客标识」字段名（兼容下划线/驼峰）。
// 这些字段会被强制改写为已通过鉴权的顾客 id，杜绝越权访问他人数据。
var customerIDFields = []string{
	"customer_id", "customerId", "customerID",
	"user_id", "userId", "userID",
}

// bindCustomerOwnership 将请求体 JSON 中的顾客标识字段强制改写为 customerID。
// 若请求体不是 JSON 或解析失败，则原样返回，依赖下游通过 X-Customer-Id 头校验归属。
func bindCustomerOwnership(body []byte, customerID int64) []byte {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var payload map[string]interface{}
	if err := dec.Decode(&payload); err != nil {
		return body
	}

	raw := strconv.FormatInt(customerID, 10)
	changed := false
	for _, field := range customerIDFields {
		v, ok := payload[field]
		if !ok {
			continue
		}
		switch val := v.(type) {
		case json.Number:
			if val.String() != raw {
				payload[field] = json.Number(raw)
				changed = true
			}
		case string:
			if val != raw {
				payload[field] = raw
				changed = true
			}
		case float64:
			if strconv.FormatInt(int64(val), 10) != raw {
				payload[field] = raw
				changed = true
			}
		}
	}

	if !changed {
		return body
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return out
}
