package customer

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/alipay"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

// NewModule 组装 customer 模块的全部依赖并返回 HTTP Handler。
//
// alipayClient 由 app 层创建后注入，为 nil 时表示支付宝未配置，
// 此时支付宝登录相关接口将返回「未配置」错误。
func NewModule(database *sqlx.DB, rdb *redis.Client, cfg *config.Config, alipayClient *alipay.Client) *Handler {
	return NewHandler(NewService(NewRepository(database), rdb, cfg, alipayClient))
}
