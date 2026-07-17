package store

// NearbyRequest 获取附近门店请求
type NearbyRequest struct {
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
	Distance  float64 `form:"distance" binding:"required,gt=0"`
	Count     int     `form:"count" binding:"required,gt=0"`
}

// GetDetailRequest 获取门店详情（含距离）请求
// id 为目标门店 ID；longitude/latitude 用于计算与门店的距离。
type GetDetailRequest struct {
	ID       int64   `form:"id" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}

// StoreDetailItem 门店信息统一响应结构，用于获取附近门店、搜索门店、获取门店详情三个接口。
// 距离展示逻辑：< 1km 以 m 展示，否则以 km 展示；distance/unit 为可选字段，
// 仅在传入经纬度时可计算，未传入时（omitempty）不返回。
type StoreDetailItem struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Status        int      `json:"status"`
	Province      string   `json:"province"`
	City          string   `json:"city"`
	District      string   `json:"district"`
	Address       string   `json:"address"`       // 完整地址，如 广东省广州市番禺区xxxxxx
	BusinessHours string   `json:"business_hours"`
	Phone         string   `json:"phone"`
	Longitude     float64  `json:"longitude"`
	Latitude      float64  `json:"latitude"`
	Distance      *float64 `json:"distance,omitempty"`
	Unit          *string  `json:"unit,omitempty"` // "m" 或 "km"
}

// SearchRequest 搜索门店请求
// city 为模糊匹配（city 为空时忽略该条件）；
// keyword 为模糊匹配（匹配门店名称或地址，keyword 为空时忽略该条件）。
// longitude/latitude 用于计算返回结果中的距离。
type SearchRequest struct {
	City      string  `form:"city"`
	Keyword   string  `form:"keyword"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}

// CityLetterGroup 城市按首字母分组响应结构
type CityLetterGroup struct {
	Letter string   `json:"letter"`
	Cities []string `json:"cities"`
}
