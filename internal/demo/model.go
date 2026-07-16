package demo

import "time"

// Product 演示用的业务实体，对应数据库中的 products 表。
//
// db tag 供 sqlx 做列名映射（NamedQuery / Get / Select 都依赖它），
// json tag 用于 HTTP 响应序列化。
type Product struct {
	ID        int64     `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"        binding:"required"`
	Price     float64   `db:"price"      json:"price"       binding:"required,gte=0"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
