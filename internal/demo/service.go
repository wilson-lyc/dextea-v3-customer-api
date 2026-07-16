package demo

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

// Service 属于「业务逻辑」层，编排 repository 与领域规则，
// 不感知 HTTP 细节，也不直接操作数据库。
type Service struct {
	repo *Repository
	cfg  *config.Config
}

// NewService 创建 Service 实例。
func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// CheckHealth 聚合应用与数据库的健康状态，返回对外响应结构。
func (s *Service) CheckHealth(ctx context.Context) (*HealthResponse, error) {
	dbStatus, _ := s.repo.Ping(ctx)

	return &HealthResponse{
		Status:  "ok",
		Service: s.cfg.ServiceName,
		DB:      dbStatus,
	}, nil
}
