package store

// NearbyRequest 获取附近门店请求
type NearbyRequest struct {
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
	Distance  float64 `form:"distance" binding:"required,gt=0"`
	Count     int     `form:"count" binding:"required,gt=0"`
}

// NearbyStoreItem 附近门店条目
type NearbyStoreItem struct {
	ID       int64   `json:"id"`       // 门店 ID
	Name     string  `json:"name"`
	Address  string  `json:"address"`  // 完整地址，如 广东省广州市番禺区xxxxxx
	Distance float64 `json:"distance"`
	Unit     string  `json:"unit"`     // "m" 或 "km"
}

// SearchRequest 搜索门店请求
// province/city/district 按省市区文本筛选（分别为空时忽略该条件）；
// keyword 模糊匹配（匹配门店名称或地址）。
// longitude/latitude 用于计算返回结果中的距离。
type SearchRequest struct {
	Province  string  `form:"province"`
	City      string  `form:"city"`
	District  string  `form:"district"`
	Keyword   string  `form:"keyword" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}
