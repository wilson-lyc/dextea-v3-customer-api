package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

const storeColumns = `id, name, province, city, district, address, status, business_hours, phone, longitude, latitude, account, password, email, created_at, updated_at`

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)

// FindByIDs 根据 ID 列表批量查询门店，返回结果保持传入 ID 的顺序
func (r *Repository) FindByIDs(ctx context.Context, ids []int64) ([]Store, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}
	if len(ids) == 0 {
		return nil, nil
	}

	q, args, err := sqlx.In(
		`SELECT `+storeColumns+` FROM stores WHERE id IN (?)`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var stores []Store
	if err := r.db.SelectContext(ctx, &stores, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return stores, nil
}

// FindByID 根据 ID 查询单个门店，未找到时返回 nil。
func (r *Repository) FindByID(ctx context.Context, id int64) (*Store, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}

	var store Store
	query := `SELECT ` + storeColumns + ` FROM stores WHERE id = ?`
	if err := r.db.GetContext(ctx, &store, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

// Search 按条件搜索门店。
// city 为文本精确匹配，匹配门店的城市字段（为空则忽略该条件）；
// keyword 为模糊匹配，匹配门店名称或地址（为空则忽略该条件）。
func (r *Repository) Search(ctx context.Context, city, keyword string) ([]Store, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}

	query := `SELECT ` + storeColumns + ` FROM stores WHERE 1=1`
	args := make([]any, 0, 2)

	if city != "" {
		query += ` AND city = ?`
		args = append(args, city)
	}
	if keyword != "" {
		// 对 % _ \ 转义，避免用户输入被当作 LIKE 通配符
		like := "%" + escapeLikeValue(keyword) + "%"
		query += ` AND (name LIKE ? ESCAPE '\\' OR address LIKE ? ESCAPE '\\')`
		args = append(args, like, like)
	}

	var stores []Store
	if err := r.db.SelectContext(ctx, &stores, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return stores, nil
}

// GetDistinctCities 查询门店表中所有去重后的城市名称，按城市名排序。
func (r *Repository) GetDistinctCities(ctx context.Context) ([]string, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}

	query := `SELECT DISTINCT city FROM stores WHERE city != '' ORDER BY city`
	var cities []string
	if err := r.db.SelectContext(ctx, &cities, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cities, nil
}

// escapeLikeValue 转义 LIKE 通配符，防止用户输入 % _ \ 干扰匹配。
func escapeLikeValue(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '%', '_', '\\':
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
