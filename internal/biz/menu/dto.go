package menu

type StoreMenuResponse struct {
	Menu   MenuInfo    `json:"menu"`
	Groups []GroupInfo `json:"groups"`
}

type MenuInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupInfo struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Sort     int           `json:"sort"`
	Products []ProductInfo `json:"products"`
}

type ProductInfo struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Brief  string  `json:"brief"`
	Price  float64 `json:"price"`
	Sort   int     `json:"sort"`
	Status int     `json:"status"`
	Cover  string  `json:"cover"`
}
