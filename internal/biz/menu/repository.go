package menu

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

// Repository 菜单模块数据访问层。
type Repository struct {
	db *sqlx.DB
}

// NewRepository 创建 Repository 实例。
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// FindStoreMenuID 根据门店 ID 查询其绑定的第一个菜单 ID。
//
// 复合主键按 store_id, menu_id 排序取首条。
// 未绑定任何菜单时返回 (0, nil)。
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

// FindMenu 根据菜单 ID 查询菜单基础信息。
//
// 未找到菜单时返回零值 Menu{}，不视为错误。
func (r *Repository) FindMenu(ctx context.Context, menuID int64) (Menu, error) {
	if r.db == nil {
		return Menu{}, bizerror.ErrMysqlDisabled
	}
	var m Menu
	err := r.db.GetContext(ctx, &m,
		`SELECT id, name, description, created_at, updated_at FROM menus WHERE id = ?`, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Menu{}, nil
		}
		return Menu{}, err
	}
	return m, nil
}

// FindGroupsByMenu 根据菜单 ID 查询其下分组。
//
// 结果按 sort 升序排列，便于上层按序渲染。
func (r *Repository) FindGroupsByMenu(ctx context.Context, menuID int64) ([]MenuGroup, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	var groups []MenuGroup
	err := r.db.SelectContext(ctx, &groups,
		`SELECT id, menu_id, name, sort, created_at, updated_at FROM menu_groups WHERE menu_id = ? ORDER BY sort`, menuID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// groupProductRow 分组商品联表查询的中间结构。
// StoreStatus 为 nullable：门店商品状态表为懒加载，未配置该商品状态时无记录，此时为 nil，
// 调用方需默认补 0（门店售罄）。ProductStatus 仅用于全局在售过滤，不作为门店状态回退值。
type groupProductRow struct {
	GroupID       int64   `db:"group_id"`
	ProductID     int64   `db:"product_id"`
	ProductSort   int     `db:"product_sort"`
	Name          string  `db:"name"`
	Brief         string  `db:"brief"`
	Price         float64 `db:"price"`
	ProductStatus int     `db:"product_status"`
	StoreStatus   *int    `db:"store_status"`
}

// productImageRow 商品封面图联表 gallery 查询的中间结构，含图片可访问地址 url。
type productImageRow struct {
	ProductID int64  `db:"product_id"`
	URL       string `db:"url"`
}

// FindGroupProducts 根据分组 ID 列表查询其下商品。
//
// 仅返回商品全局状态为在售（products.status = 1）的商品，并通过 LEFT JOIN 获取门店商品状态。
// 门店商品状态表为懒加载：若未配置则返回 nil，调用方需默认补 0（售罄）。
//
// 结果按 group_id、product_sort 升序排列。
func (r *Repository) FindGroupProducts(ctx context.Context, groupIDs []int64, storeID int64) ([]groupProductRow, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(groupIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT mp.group_id AS group_id,
		       mp.product_id AS product_id,
		       mp.sort AS product_sort,
		       p.name AS name,
		       p.brief AS brief,
		       p.price AS price,
		       p.status AS product_status,
		       pss.status AS store_status
		FROM menu_products mp
		JOIN products p ON p.id = mp.product_id AND p.status = 1
		LEFT JOIN product_store_status pss ON pss.product_id = p.id AND pss.store_id = ?
		WHERE mp.group_id IN (?)`, storeID, groupIDs)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var rows []groupProductRow
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// FindProductImages 根据商品 ID 列表批量查询商品封面图。
//
// 每个商品取 type=1 的第一张图片（按 MIN(id) 确定），通过 LEFT JOIN gallery 补上可访问地址。
func (r *Repository) FindProductImages(ctx context.Context, productIDs []int64) ([]productImageRow, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}
	if len(productIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT pi.product_id AS product_id,
		       g.url AS url
		FROM product_images pi
		LEFT JOIN gallery g ON g.id = pi.image_id
		WHERE pi.type = 1
		  AND pi.product_id IN (?)
		  AND pi.image_id = (
			SELECT pi2.image_id FROM product_images pi2
			WHERE pi2.product_id = pi.product_id AND pi2.type = 1
			ORDER BY pi2.sort ASC, pi2.image_id ASC
			LIMIT 1
		  )`,
		productIDs,
	)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)

	var imgs []productImageRow
	if err := r.db.SelectContext(ctx, &imgs, q, args...); err != nil {
		return nil, err
	}
	return imgs, nil
}
