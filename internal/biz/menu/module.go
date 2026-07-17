package menu

import "github.com/jmoiron/sqlx"

// NewModule 组装 menu 模块的全部依赖并返回 HTTP Handler。
func NewModule(database *sqlx.DB) *Handler {
	return NewHandler(NewService(NewRepository(database)))
}
