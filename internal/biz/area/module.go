package area

import (
	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/storerpc"
)

// NewModule 组装 area 模块的全部依赖并返回 HTTP Handler。
func NewModule(client *storerpc.Client, rdb *redis.Client, cfg *config.Config) *Handler {
	return NewHandler(NewService(cfg, client, rdb))
}
