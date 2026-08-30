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

const maxRetries = 2

var orderBreaker = newCircuitBreaker(5, 10*time.Second)

type ForwardResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type Service struct {
	httpClient *http.Client
	baseURL    string
	resolver   *nacos.Resolver
}

func NewService(cfg *config.Config) *Service {
	svc := &Service{
		httpClient: &http.Client{Timeout: orderHTTPTimeout},
		baseURL:    strings.TrimRight(cfg.OrderServiceBaseURL, "/"),
	}
	if cfg.OrderServiceMode == "nacos" {
		namingClient, err := nacos.NewNamingClient(cfg.NacosConfig())
		if err == nil {
			svc.resolver = nacos.NewResolver(namingClient, cfg.OrderServiceName, cfg.OrderServiceGroup)
		}
	}
	return svc
}

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

func (s *Service) Forward(ctx context.Context, customerID int64, method, path, rawQuery string, body []byte) (*ForwardResult, error) {
	baseURL, err := s.resolveBaseURL()
	if err != nil {
		return nil, err
	}

	target := baseURL + buildDownstreamPath(path)
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
	req.Header.Set(consts.CustomerIDHeader, strconv.FormatInt(customerID, 10))

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	result, err := s.doForward(req)
	if err != nil {
		orderBreaker.recordFailure()
		return nil, err
	}
	orderBreaker.recordSuccess()
	return result, nil
}

func (s *Service) doForward(req *http.Request) (*ForwardResult, error) {
	operation := func() (*ForwardResult, error) {
		resp, err := s.httpClient.Do(req)
		if err != nil {
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

func buildDownstreamPath(localPath string) string {
	localPrefix := consts.APIPrefixV1 + "/orders"
	rel := strings.TrimPrefix(localPath, localPrefix)
	if rel == "" {
		rel = "/"
	}
	return consts.DownstreamAPIPrefix + rel
}
