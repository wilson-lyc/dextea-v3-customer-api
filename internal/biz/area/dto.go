package area

// ReverseGeocodeRequest 逆地址编码请求
type ReverseGeocodeRequest struct {
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
}

// ReverseGeocodeResponse 逆地址编码响应（仅返回省市区名称）
type ReverseGeocodeResponse struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

// CityLetterGroup 城市按首字母分组响应结构
type CityLetterGroup struct {
	Letter string   `json:"letter"`
	Cities []string `json:"cities"`
}

// amapRegeoResponse 高德逆地址编码 API 原始响应
type amapRegeoResponse struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Regeocode struct {
		AddressComponent struct {
			Province string `json:"province"`
			City     string `json:"city"`
			District string `json:"district"`
		} `json:"addressComponent"`
	} `json:"regeocode"`
}
