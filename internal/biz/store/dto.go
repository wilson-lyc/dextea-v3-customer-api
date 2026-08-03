package store

type NearbyRequest struct {
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
	Distance  float64 `form:"distance" binding:"required,gt=0"`
	Count     int     `form:"count" binding:"required,gt=0"`
}

type GetDetailRequest struct {
	ID        int64   `form:"id" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}

type StoreDetailItem struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Status        int      `json:"status"`
	Province      string   `json:"province"`
	City          string   `json:"city"`
	District      string   `json:"district"`
	Address       string   `json:"address"`
	BusinessHours string   `json:"business_hours"`
	Phone         string   `json:"phone"`
	Longitude     float64  `json:"longitude"`
	Latitude      float64  `json:"latitude"`
	Distance      *float64 `json:"distance,omitempty"`
	Unit          *string  `json:"unit,omitempty"` // m 或 km
}

type SearchRequest struct {
	City      string  `form:"city"`    // 完全匹配
	Keyword   string  `form:"keyword"` // 模糊匹配
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}
