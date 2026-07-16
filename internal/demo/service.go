package demo

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/redis/go-redis/v9"
)

// Service 属于「业务逻辑」层，编排 repository 与领域规则，
// 不感知 HTTP 细节，也不直接操作数据库。
type Service struct {
	repo *Repository
	rdb  *redis.Client // Redis 客户端；为 nil 时表示未启用。
	cfg  *config.Config
}

// NewService 创建 Service 实例。
func NewService(repo *Repository, rdb *redis.Client, cfg *config.Config) *Service {
	return &Service{repo: repo, rdb: rdb, cfg: cfg}
}

// CheckHealth 聚合应用、MySQL 与 Redis 的健康状态，返回对外响应结构。
func (s *Service) CheckHealth(ctx context.Context) (*HealthResponse, error) {
	dbStatus, _ := s.repo.Ping(ctx)
	redisStatus := s.pingRedis(ctx)

	return &HealthResponse{
		Status:  "ok",
		Service: s.cfg.ServiceName,
		DB:      dbStatus,
		Redis:   redisStatus,
	}, nil
}

// pingRedis 检查 Redis 连通性，返回状态字符串：ok / down / disabled。
func (s *Service) pingRedis(ctx context.Context) string {
	if s.rdb == nil {
		return "disabled"
	}
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return "down"
	}
	return "ok"
}

// CreateProduct 创建商品。
func (s *Service) CreateProduct(ctx context.Context, in *Product) (*Product, error) {
	if s.repo == nil {
		return nil, ErrDBDisabled
	}
	if err := s.repo.Create(ctx, in); err != nil {
		return nil, err
	}
	return in, nil
}

// ErrDBDisabled 表示未配置数据库时的业务异常。
var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)
