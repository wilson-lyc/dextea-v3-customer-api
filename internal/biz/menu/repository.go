package menu

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

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)

// FindStoreMenuID 根据门店 ID 查询其绑定的第一个菜单 ID（复合主键按 store_id, menu_id 排序取首条）。
// 未绑定任何菜单时返回 (0, nil)。
func (r *Repository) FindStoreMenuID(ctx context.Context, storeID int64) (int64, error) {
	if r.db == nil {
		return 0, ErrDBDisabled
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
func (r *Repository) FindMenu(ctx context.Context, menuID int64) (Menu, error) {
	if r.db == nil {
		return Menu{}, ErrDBDisabled
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

// FindGroupsByMenu 根据菜单 ID 查询其下分组，按 sort 升序返回。
func (r *Repository) FindGroupsByMenu(ctx context.Context, menuID int64) ([]MenuGroup, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
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
// StoreStatus 为 nullable：门店未配置该商品状态时为 nil，调用方回退到 ProductStatus。
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

// productImageRow 商品图片联表 gallery 查询的中间结构，含图片可访问地址 url。
type productImageRow struct {
	ProductID int64  `db:"product_id"`
	ImageID   int64  `db:"image_id"`
	Type      int    `db:"type"`
	Sort      int    `db:"sort"`
	URL       string `db:"url"`
}

// FindGroupProducts 根据分组 ID 列表查询其下商品。
// 仅返回商品全局状态（products.status）为 1 的商品；并通过 LEFT JOIN 取门店商品状态（product_store_status）。
// 结果按 group_id、分组内商品排序（product_sort）升序，便于上层按序分组。
func (r *Repository) FindGroupProducts(ctx context.Context, groupIDs []int64, storeID int64) ([]groupProductRow, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
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

// FindProductImages 根据商品 ID 列表批量查询商品主图，仅取 type=1 的图片，
// 并通过 LEFT JOIN gallery 补上图片可访问地址（url）。结果按 product_id、sort 升序，
// 便于上层对每个商品取命中排序最小的第一张作为主图。
func (r *Repository) FindProductImages(ctx context.Context, productIDs []int64) ([]productImageRow, error) {
	if r.db == nil {
		return nil, ErrDBDisabled
	}
	if len(productIDs) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`
		SELECT pi.product_id AS product_id,
		       pi.image_id AS image_id,
		       pi.type AS type,
		       pi.sort AS sort,
		       g.url AS url
		FROM product_images pi
		LEFT JOIN gallery g ON g.id = pi.image_id
		WHERE pi.product_id IN (?) AND pi.type = 1 ORDER BY pi.product_id, pi.sort`,
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
