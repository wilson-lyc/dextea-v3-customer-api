package demo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Repository 属于「数据访问」层，专门负责与数据库的连接与 SQL 操作。
// 上层（service）只调用语义化方法，不直接碰 sqlx.DB。
type Repository struct {
	db *sqlx.DB
}

// NewRepository 创建 Repository 实例。db 为 nil 时表示未配置数据库。
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Ping 检查数据库连通性，返回状态字符串：ok / down / disabled。
// 这里演示的是「探针」类操作；后续真实的增删改查方法也放在本文件，
// 借助 sqlx 的 GetContext / SelectContext 等做结构体映射。
func (r *Repository) Ping(ctx context.Context) (string, error) {
	if r.db == nil {
		return "disabled", nil
	}
	if err := r.db.PingContext(ctx); err != nil {
		return "down", err
	}
	return "ok", nil
}
