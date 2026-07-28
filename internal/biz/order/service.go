package order

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"net/url"
	"strings"
	"time"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/middleware"
)

const orderHTTPTimeout = 10 * time.Second

type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type Service struct {
	httpClient   *http.Client
	baseURL      string
	createPath   string
	preBuildPath string
	listPath     string
	detailPath   string
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		httpClient:   &http.Client{Timeout: orderHTTPTimeout},
		baseURL:      strings.TrimRight(cfg.OrderServiceBaseURL, "/"),
		createPath:   cfg.OrderCreatePath,
		preBuildPath: cfg.OrderPreBuildPath,
		listPath:     cfg.OrderListPath,
		detailPath:   cfg.OrderDetailPath,
	}
}

func (s *Service) Create(ctx context.Context, customerID int64, body []byte) (*ForwardResult, error) {
	return s.forward(ctx, s.createPath, customerID, body)
}

func (s *Service) PreBuild(ctx context.Context, customerID int64, body []byte) (*ForwardResult, error) {
	return s.forward(ctx, s.preBuildPath, customerID, body)
}

// customerIDFields 是请求体中可能出现的「顾客标识」字段名（兼容下划线/驼峰）。
// 这些字段会被强制改写为已通过鉴权的顾客 id，杜绝越权访问他人数据。
var customerIDFields = []string{
	"customer_id", "customerId", "customerID",
	"user_id", "userId", "userID",
}

func (s *Service) forward(ctx context.Context, path string, customerID int64, body []byte) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(CodeOrderServiceNotConfigured)
	}

	url := s.baseURL + path
	// 用已认证的顾客 id 覆盖请求体中的顾客标识，防止 A 的令牌操作/查询 B 的数据。
	body = bindCustomerOwnership(body, customerID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceError, "构建订单服务请求失败")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// 透传已认证的顾客身份，供下游订单服务据此做数据归属校验。
	httpReq.Header.Set(middleware.CustomerIDHeader, strconv.FormatInt(customerID, 10))

	resp, err := s.httpClient.Do(httpReq)
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
