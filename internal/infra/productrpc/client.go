package productrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	productv1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/nacos"
)

var (
	ErrUnavailable = &bizerror.BizError{Code: 43001, Message: "商品服务暂不可用", Kind: bizerror.KindDownstream}
	ErrNotFound    = &bizerror.BizError{Code: 40400, Message: "资源不存在", Kind: bizerror.KindBusiness}
)

// Client 负责商品服务的寻址、连接和统一错误映射。连接按请求创建，避免长期持有失效的 Nacos 实例。
type Client struct {
	baseURL  string
	resolver *nacos.Resolver
}

func NewClient(cfg *config.Config) (*Client, error) {
	c := &Client{baseURL: strings.TrimSpace(cfg.ProductServiceBaseURL)}
	if cfg.ProductServiceMode == "nacos" {
		naming, err := nacos.NewNamingClient(cfg.NacosConfig())
		if err != nil {
			return nil, fmt.Errorf("init product service discovery: %w", err)
		}
		c.resolver = nacos.NewResolver(naming, cfg.ProductServiceName, cfg.ProductServiceGroup)
	}
	return c, nil
}

func (c *Client) withClient(ctx context.Context, fn func(productv1.ProductServiceClient) error) error {
	addr := c.baseURL
	if c.resolver != nil {
		resolved, err := c.resolver.Resolve()
		if err != nil {
			return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
		}
		addr = resolved
	}
	if addr == "" {
		return bizerror.New(ErrUnavailable)
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4<<20)))
	if err != nil {
		return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
	}
	defer conn.Close()
	if err := fn(productv1.NewProductServiceClient(conn)); err != nil {
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
		return bizerror.NewWith(&bizerror.BizError{Code: 40001, Message: "请求参数不合法", Kind: bizerror.KindBusiness}, bizerror.WithCause(err))
	default:
		return bizerror.NewWith(ErrUnavailable, bizerror.WithCause(err))
	}
}

func (c *Client) Detail(ctx context.Context, req *productv1.GetProductDetailRequest) (*productv1.ProductDetail, error) {
	var out *productv1.ProductDetail
	err := c.withClient(ctx, func(client productv1.ProductServiceClient) error {
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var err error
		out, err = client.GetProductDetail(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) StoreStatuses(ctx context.Context, req *productv1.GetProductStoreStatusesRequest) (*productv1.ProductStoreStatusesResponse, error) {
	var out *productv1.ProductStoreStatusesResponse
	err := c.withClient(ctx, func(client productv1.ProductServiceClient) error {
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var err error
		out, err = client.GetProductStoreStatuses(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) MenuTree(ctx context.Context, req *productv1.GetMenuTreeRequest) (*productv1.MenuTreeResponse, error) {
	var out *productv1.MenuTreeResponse
	err := c.withClient(ctx, func(client productv1.ProductServiceClient) error {
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var err error
		out, err = client.GetMenuTree(callCtx, req)
		return err
	})
	return out, err
}

func (c *Client) Menu(ctx context.Context, req *productv1.GetMenuRequest) (*productv1.Menu, error) {
	var out *productv1.Menu
	err := c.withClient(ctx, func(client productv1.ProductServiceClient) error {
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var err error
		out, err = client.GetMenu(callCtx, req)
		return err
	})
	return out, err
}
