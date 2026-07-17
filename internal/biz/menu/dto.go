package menu

// StoreMenuResponse 门店菜单返回结构：一个门店对应一个当前菜单，菜单下包含多个分组，
// 每个分组下包含若干商品（仅返回全局在售商品，状态跟随门店状态）。
type StoreMenuResponse struct {
	Menu   MenuInfo    `json:"menu"`
	Groups []GroupInfo `json:"groups"`
}

// MenuInfo 菜单基础信息
type MenuInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GroupInfo 菜单分组及旗下商品
type GroupInfo struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Sort     int           `json:"sort"`
	Products []ProductInfo `json:"products"`
}

// ProductInfo 返回给前端的商品数据：名称、简介、图片、价格，状态取自门店商品状态。
type ProductInfo struct {
	ID     int64              `json:"id"`
	Name   string             `json:"name"`
	Brief  string             `json:"brief"`
	Price  float64            `json:"price"`
	Status int                `json:"status"` // 门店商品状态（取自 product_store_status，缺省回退全局状态）
	Images []ProductImageInfo `json:"images"`
}

// ProductImageInfo 商品图片（image_id 指向图库，url 为可访问地址，type 区分图片类型，sort 为排序）
type ProductImageInfo struct {
	ImageID int64  `json:"image_id"`
	URL     string `json:"url"`
	Type    int    `json:"type"`
	Sort    int    `json:"sort"`
}
