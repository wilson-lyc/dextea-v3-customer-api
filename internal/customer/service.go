package customer

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo *Repository
	rdb  *redis.Client
	cfg  *config.Config
}

func NewService(repo *Repository, rdb *redis.Client, cfg *config.Config) *Service {
	return &Service{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *Service) Login(ctx context.Context, code string, platform Platform) (*LoginResponse, error) {
	return &LoginResponse{Code: code}, nil
}

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)
