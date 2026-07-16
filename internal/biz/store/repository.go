package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

const storeColumns = `id, name, region_code, address, status, business_hours, phone, longitude, latitude, account, password, email, created_at, updated_at`

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
