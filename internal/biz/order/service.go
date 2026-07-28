package order

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
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

func (s *Service) Create(ctx context.Context, body []byte) (*ForwardResult, error) {
	return s.forward(ctx, s.createPath, body)
}

func (s *Service) PreBuild(ctx context.Context, body []byte) (*ForwardResult, error) {
	return s.forward(ctx, s.preBuildPath, body)
}

// List 转发订单列表查询（GET）。
func (s *Service) List(ctx context.Context, rawQuery string) (*ForwardResult, error) {
	return s.forwardGet(ctx, s.listPath, rawQuery)
}

// Detail 转发订单详情查询（GET /{orderId}）。
func (s *Service) Detail(ctx context.Context, orderID string, rawQuery string) (*ForwardResult, error) {
	path := s.detailPath + "/" + url.PathEscape(orderID)
	return s.forwardGet(ctx, path, rawQuery)
}

func (s *Service) forward(ctx context.Context, path string, body []byte) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(CodeOrderServiceNotConfigured)
	}

	url := s.baseURL + path
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceError, "构建订单服务请求失败")
	}
	httpReq.Header.Set("Content-Type", "application/json")

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

// forwardGet 转发 GET 请求到订单服务，保留原始查询参数。
func (s *Service) forwardGet(ctx context.Context, path string, rawQuery string) (*ForwardResult, error) {
	if s.baseURL == "" {
		return nil, bizerror.New(CodeOrderServiceNotConfigured)
	}

	u := s.baseURL + path
	if rawQuery != "" {
		u += "?" + rawQuery
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, bizerror.New(CodeOrderServiceError, "构建订单服务请求失败")
	}

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
