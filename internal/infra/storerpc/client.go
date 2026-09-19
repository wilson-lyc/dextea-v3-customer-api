package storerpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/nacos"
)

var (
	ErrUnavailable = &bizerror.BizError{Code: 43002, Message: "门店服务暂不可用", Kind: bizerror.KindDownstream}
	ErrNotFound    = &bizerror.BizError{Code: 40400, Message: "资源不存在", Kind: bizerror.KindBusiness}
)

// Client 只暴露 StoreBusinessService，调用方不能意外获得管理面或凭证面权限。
type Client struct {
	baseURL  string
	token    string
	resolver *nacos.Resolver
}

func NewClient(cfg *config.Config) (*Client, error) {
	c := &Client{
		baseURL: strings.TrimSpace(cfg.StoreServiceBaseURL),
		token:   strings.TrimSpace(cfg.StoreServiceToken),
	}
	if cfg.StoreServiceMode == "nacos" {
		naming, err := nacos.NewNamingClient(cfg.NacosConfig())
		if err != nil {
			if c.baseURL == "" {
				return nil, fmt.Errorf("init store service discovery: %w", err)
			}
			return c, nil
		}
		c.resolver = nacos.NewResolver(naming, cfg.StoreServiceName, cfg.StoreServiceGroup)
	}
	return c, nil
}

func (c *Client) withClient(ctx context.Context, fn func(storev1.StoreBusinessServiceClient, context.Context) error) error {
	addr := c.baseURL
	if c.resolver != nil {
		resolved, err := c.resolver.Resolve()
		if err == nil {
			addr = resolved
		} else if addr == "" {
			return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
		}
	}
	if addr == "" {
		return bizerror.New(ErrUnavailable)
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4<<20)),
	)
	if err != nil {
		return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
	}
	defer conn.Close()

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if c.token != "" {
		callCtx = metadata.AppendToOutgoingContext(callCtx, "x-service-token", c.token)
	}
	if err := fn(storev1.NewStoreBusinessServiceClient(conn), callCtx); err != nil {
		return mapError(err)
	}
	return nil
}

func mapError(err error) error {
	if _, ok := bizerror.As(err); ok {
		return err
	}
	switch status.Code(err) {
	case codes.NotFound:
		return bizerror.NewWith(ErrNotFound, bizerror.WithCause(err))
	case codes.InvalidArgument:
		return bizerror.NewWith(bizerror.ErrBadRequest, bizerror.WithCause(err))
	default:
		return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
	}
}

func (c *Client) Detail(ctx context.Context, req *storev1.GetBusinessStoreRequest) (*storev1.BusinessStoreDistance, error) {
	var out *storev1.BusinessStoreDistance
	err := c.withClient(ctx, func(client storev1.StoreBusinessServiceClient, callCtx context.Context) error {
		var err error
		out, err = client.GetStore(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) Batch(ctx context.Context, req *storev1.GetBusinessStoresRequest) (*storev1.GetBusinessStoresResponse, error) {
	var out *storev1.GetBusinessStoresResponse
	err := c.withClient(ctx, func(client storev1.StoreBusinessServiceClient, callCtx context.Context) error {
		var err error
		out, err = client.GetStores(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) Search(ctx context.Context, req *storev1.SearchBusinessStoresRequest) (*storev1.SearchBusinessStoresResponse, error) {
	var out *storev1.SearchBusinessStoresResponse
	err := c.withClient(ctx, func(client storev1.StoreBusinessServiceClient, callCtx context.Context) error {
		var err error
		out, err = client.SearchStores(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) Cities(ctx context.Context) (*storev1.ListStoreCitiesResponse, error) {
	var out *storev1.ListStoreCitiesResponse
	err := c.withClient(ctx, func(client storev1.StoreBusinessServiceClient, callCtx context.Context) error {
		var err error
		out, err = client.ListStoreCities(callCtx, &storev1.ListStoreCitiesRequest{})
		return err
	})
	return out, err
}

func (c *Client) Nearby(ctx context.Context, req *storev1.GetBusinessNearbyStoresRequest) (*storev1.GetBusinessNearbyStoresResponse, error) {
	var out *storev1.GetBusinessNearbyStoresResponse
	err := c.withClient(ctx, func(client storev1.StoreBusinessServiceClient, callCtx context.Context) error {
		var err error
		out, err = client.GetNearbyStores(callCtx, req)
		return err
	})
	return out, err
}
