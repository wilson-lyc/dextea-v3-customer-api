package order

import (
	"bytes"
	"context"
	"io"
	"net/http"
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
	httpClient    *http.Client
	baseURL       string
	calculatePath string
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		httpClient:    &http.Client{Timeout: orderHTTPTimeout},
		baseURL:       strings.TrimRight(cfg.OrderServiceBaseURL, "/"),
		calculatePath: cfg.OrderCalculatePath,
	}
}

func (s *Service) Calculate(ctx context.Context, body []byte) (*ForwardResult, error) {
	return s.forward(ctx, s.calculatePath, body)
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
