package product

type GetProductDetailRequest struct {
	ProductID int64 `form:"productId" binding:"required"`
	StoreID   int64 `form:"storeId" binding:"required"`
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
