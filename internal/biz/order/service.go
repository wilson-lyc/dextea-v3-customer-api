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

// goAPIPrefix 是 Go 服务对外挂载 Order 模块的 API 前缀。
// 转发时需将该前缀从原始请求路径中剥离，再拼接到下游基础地址之后，
// 否则会与 ORDER_SERVICE_BASE_URL 中已包含的相同前缀重复。
const goAPIPrefix = "/api/v1"

// ForwardResult 封装下游订单服务返回的原始响应，由 handler 透传写出。
type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Service 负责把 Order 模块的请求转发到下游 Java 订单服务，自身不执行业务逻辑。
// 转发是「目标地址无关」的：不写死任何下游接口路径，请求的 method / path /
// query / body 原样透传，仅完成统一的身份注入与请求体清洗。
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

// Forward 通用转发：把进入 Order 模块的请求原样转发到下游订单服务。
//
// path 为原始请求路径（含 Go 服务自身前缀，如 /api/v1/orders/123），转发时会
// 剥离 /api/v1 前缀后再拼接 baseURL；method、rawQuery、body 原样透传。
// 请求体中的顾客标识字段会被强制改写为已认证的 customerID，顾客身份通过
// 固定头 X-Customer-Id 透传，供下游据此做数据归属校验。
func (s *Service) Forward(ctx context.Context, customerID int64, method, path, rawQuery string, body []byte) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(ErrOrderServiceNotConfigured)
	}

	target := s.baseURL + stripAPIPrefix(path)
	if rawQuery != "" {
		target += "?" + rawQuery
	}

	// 清洗：仅当存在请求体时，用已认证的顾客 id 覆盖请求体中的顾客标识字段，
	// 杜绝 A 的令牌操作 B 的数据。
	var reader io.Reader
	if len(body) > 0 {
		body = bindCustomerOwnership(body, customerID)
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
	req.Header.Set(CustomerIDHeader, strconv.FormatInt(customerID, 10))

	return doForward(s.httpClient, req)
}

// stripAPIPrefix 剥离请求路径中 Go 服务自身的 API 前缀，避免与下游 baseURL 重复。
// 仅当路径等于该前缀或其后的下一段紧跟 "/" 时才剥离（如 /api/v1/orders），
// 避免误伤 /api/v10 这类以相同前缀开头的路径；剥离结果为空时回退为 "/"。
func stripAPIPrefix(path string) string {
	p := path
	switch {
	case p == goAPIPrefix:
		return "/"
	case strings.HasPrefix(p, goAPIPrefix+"/"):
		p = strings.TrimPrefix(p, goAPIPrefix)
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
