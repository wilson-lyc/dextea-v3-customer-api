package customer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

const customerColumns = `id, name, COALESCE(weixin_open_id, '') AS weixin_open_id, COALESCE(alipay_open_id, '') AS alipay_open_id, COALESCE(email, '') AS email, COALESCE(phone, '') AS phone, COALESCE(password, '') AS password, status, created_at, updated_at`

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *Repository) FindByAlipayOpenID(ctx context.Context, openID string) (*Customer, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
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

func (r *Repository) FindByWeixinOpenID(ctx context.Context, openID string) (*Customer, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	const q = `SELECT ` + customerColumns + ` FROM customers WHERE weixin_open_id = ? LIMIT 1`
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
		return nil, bizerror.ErrMysqlDisabled
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
		return nil, bizerror.ErrMysqlDisabled
	}
	const q = `INSERT INTO customers (name, weixin_open_id, alipay_open_id, email, phone, password, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q,
		c.Name,
		nullableString(c.WeixinOpenID),
		nullableString(c.AlipayOpenID),
		nullableString(c.Email),
		nullableString(c.Phone),
		nullableString(c.Password),
		c.Status,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}
