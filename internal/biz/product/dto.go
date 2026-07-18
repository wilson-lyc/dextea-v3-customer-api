package product

type GetProductDetailRequest struct {
	ProductID int64 `form:"productId" binding:"required"`
	StoreID   int64 `form:"storeId" binding:"required"`
}

// GetProductStoreStatusRequest 获取门店下商品状态请求
// productId 选填：
//   - 不填：返回全局上架（status=1）商品在该门店的状态列表。
//   - 填写：仅返回指定商品在该门店的状态，且该商品必须为全局上架，否则返回业务错误。
type GetProductStoreStatusRequest struct {
	StoreID   int64 `form:"storeId" binding:"required"`
	ProductID int64 `form:"productId"`
}

// ProductStoreStatusItem 单个商品在指定门店下的状态
type ProductStoreStatusItem struct {
	ProductID int64  `json:"productId"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
}

type ProductDetailResponse struct {
	ID             int64                  `json:"id"`
	Name           string                 `json:"name"`
	Brief          string                 `json:"brief"`
	Description    string                 `json:"description"`
	Status         int                    `json:"status"`
	Price          float64                `json:"price"`
	Images         []ProductImageItem     `json:"images"`
	Customizations []CustomizationItem    `json:"customizations"`
}

type ProductImageItem struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
	Sort    int    `json:"sort"`
	Type    int    `json:"type"`
}

type CustomizationItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	Sort     int                       `json:"sort"`
	Status   int                       `json:"status"`
	Options  []CustomizationOptionItem `json:"options"`
}

type CustomizationOptionItem struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	Sort    int     `json:"sort"`
	Status  int     `json:"status"`
}
