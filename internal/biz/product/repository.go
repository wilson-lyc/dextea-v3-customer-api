package product

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

const (
	productColumns                        = `id, name, brief, description, status, price`
	customizationItemColumns              = `id, product_id, name, sort, status, created_at, updated_at`
	customizationOptionColumns            = `id, item_id, name, price, sort, status, ingredient_id, ingredient_quantity, created_at, updated_at`
	customizationOptionStoreStatusColumns = `option_id, store_id, status, created_at, updated_at`
)

func (r *Repository) FindByID(ctx context.Context, productID int64) (*Product, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	var p Product
	query := `SELECT ` + productColumns + ` FROM products WHERE id = ?`
	if err := r.db.GetContext(ctx, &p, query, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) FindProductStoreStatus(ctx context.Context, productID, storeID int64) (*int, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	var status int
	query := `SELECT status FROM product_store_status WHERE product_id = ? AND store_id = ?`
	if err := r.db.GetContext(ctx, &status, query, productID, storeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *Repository) FindActiveProducts(ctx context.Context) ([]Product, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	query := `SELECT id, name FROM products WHERE status = 1 ORDER BY id`
	var list []Product
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (r *Repository) FindByIDs(ctx context.Context, productIDs []int64) ([]Product, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(productIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT id, name FROM products WHERE id IN (?) AND status = 1 ORDER BY id`,
		productIDs)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	var list []Product
	if err := r.db.SelectContext(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (r *Repository) FindStoreStatuses(ctx context.Context, productIDs []int64, storeID int64) (map[int64]int, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(productIDs) == 0 {
		return map[int64]int{}, nil
	}

	q, args, err := sqlx.In(`
		SELECT product_id, status FROM product_store_status
		WHERE product_id IN (?) AND store_id = ?`,
		productIDs, storeID)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var rows []struct {
		ProductID int64 `db:"product_id"`
		Status    int   `db:"status"`
	}
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}

	m := make(map[int64]int, len(rows))
	for _, row := range rows {
		m[row.ProductID] = row.Status
	}
	return m, nil
}

func (r *Repository) FindCustomizationItems(ctx context.Context, productID int64) ([]CustomizationItem, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	var list []CustomizationItem
	query := `SELECT ` + customizationItemColumns + `
		FROM customization_items WHERE product_id = ? AND status = 1 ORDER BY sort`
	if err := r.db.SelectContext(ctx, &list, query, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (r *Repository) FindCustomizationOptions(ctx context.Context, itemIDs []int64) ([]CustomizationOption, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(itemIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT `+customizationOptionColumns+`
		FROM customization_options WHERE item_id IN (?) AND status = 1
		ORDER BY item_id, sort`, itemIDs)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var list []CustomizationOption
	if err := r.db.SelectContext(ctx, &list, q, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) FindCustomizationOptionStoreStatuses(ctx context.Context, optionIDs []int64, storeID int64) ([]CustomizationOptionStoreStatus, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(optionIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT `+customizationOptionStoreStatusColumns+`
		FROM customization_option_store_status WHERE option_id IN (?) AND store_id = ?`,
		optionIDs, storeID)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var list []CustomizationOptionStoreStatus
	if err := r.db.SelectContext(ctx, &list, q, args...); err != nil {
		return nil, err
	}
	return list, nil
}

type productImageRow struct {
	ImageID int64  `db:"image_id"`
	URL     string `db:"url"`
	Sort    int    `db:"sort"`
	Type    int    `db:"type"`
}

func (r *Repository) FindProductImages(ctx context.Context, productID int64) ([]productImageRow, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	query := `
		SELECT pi.image_id AS image_id, g.url AS url, pi.sort AS sort, pi.type AS type
		FROM product_images pi
		LEFT JOIN gallery g ON g.id = pi.image_id
		WHERE pi.product_id = ? AND pi.type IN (1, 2) ORDER BY pi.type, pi.sort`
	var list []productImageRow
	if err := r.db.SelectContext(ctx, &list, query, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}
