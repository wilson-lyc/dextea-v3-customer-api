package menu

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

// Repository 只保留门店域的菜单绑定查询；菜单及其商品树归属商品服务。
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// FindStoreMenuID 根据门店 ID 查询其绑定的第一个菜单 ID。
func (r *Repository) FindStoreMenuID(ctx context.Context, storeID int64) (int64, error) {
	if r.db == nil {
		return 0, bizerror.ErrMysqlDisabled
	}
	var menuID int64
	err := r.db.GetContext(ctx, &menuID,
		`SELECT menu_id FROM store_menus WHERE store_id = ? ORDER BY store_id, menu_id LIMIT 1`, storeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return menuID, nil
}
