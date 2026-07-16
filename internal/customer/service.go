package customer

import (
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

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)
