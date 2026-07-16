package demo

import (
	"context"
	"time"

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
func (r *Repository) Ping(ctx context.Context) (string, error) {
	if r.db == nil {
		return "disabled", nil
	}
	if err := r.db.PingContext(ctx); err != nil {
		return "down", err
	}
	return "ok", nil
}

// Migrate 建表（幂等）。使用 IF NOT EXISTS，重复调用安全。
// 真实项目中通常交给 flyway / golang-migrate 等迁移工具，这里仅为演示。
//
// 表结构按 MySQL 语法编写：自增主键用 BIGINT AUTO_INCREMENT，金额用 DECIMAL 以避免浮点误差。
func (r *Repository) Migrate(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	const ddl = `
		CREATE TABLE IF NOT EXISTS products (
			id         BIGINT PRIMARY KEY AUTO_INCREMENT,
			name       VARCHAR(255) NOT NULL,
			price      DECIMAL(12,2) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

// Create 插入一条商品记录，并通过 LastInsertId 回填数据库生成的自增主键。
//
// MySQL 不支持 INSERT ... RETURNING 语法，因此用 NamedExecContext 执行插入，
// 再读取 Result.LastInsertId() 获取自增 id；created_at / updated_at 已在方法内显式赋值，
// 无需从数据库回读（若表列设置了默认值，也可省略赋值）。
func (r *Repository) Create(ctx context.Context, p *Product) error {
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
		INSERT INTO products (name, price, created_at, updated_at)
		VALUES (:name, :price, :created_at, :updated_at);
	`
	res, err := r.db.NamedExecContext(ctx, query, p)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}
