package customer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

const customerColumns = `id, name, weixin_open_id, alipay_open_id, email, phone, password, status, platform, created_at, updated_at`

func (r *Repository) FindByAlipayOpenID(ctx context.Context, openID string) (*Customer, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}
	const q = `SELECT ` + customerColumns + ` FROM customers WHERE alipay_open_id = ? LIMIT 1`
	var c Customer
	err := r.db.GetContext(ctx, &c, q, openID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Customer, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}
	const q = `SELECT ` + customerColumns + ` FROM customers WHERE id = ? LIMIT 1`
	var c Customer
	if err := r.db.GetContext(ctx, &c, q, id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) Create(ctx context.Context, c *Customer) (*Customer, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}
	const q = `INSERT INTO customers (name, alipay_open_id, status, platform) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, c.Name, c.AlipayOpenID, c.Status, c.Platform)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}
