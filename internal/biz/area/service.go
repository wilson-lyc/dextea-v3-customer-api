package area

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

const (
	amapRegeoURL    = "https://restapi.amap.com/v3/geocode/regeo"
	amapHTTPTimeout = 5 * time.Second
)

type Service struct {
	httpClient *http.Client
	apiKey     string
}

func NewService(apiKey string) *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: amapHTTPTimeout,
		},
		apiKey: apiKey,
	}
}

// ReverseGeocode 调用高德逆地址编码 API，返回省市区名称。
// 高德返回什么，我们就返回什么。
func (s *Service) ReverseGeocode(ctx context.Context, req ReverseGeocodeRequest) (*ReverseGeocodeResponse, error) {
	// 参数校验
	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, bizerror.New(CodeInvalidLocation, "经度取值范围为 -180 到 180")
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, bizerror.New(CodeInvalidLocation, "纬度取值范围为 -90 到 90")
	}

	// 构造请求 URL：经度在前，纬度在后
	url := fmt.Sprintf("%s?location=%f,%f&key=%s", amapRegeoURL, req.Longitude, req.Latitude, s.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create amap request: %w", err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, bizerror.New(CodeAmapUnavailable, "请求高德地图服务失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read amap response: %w", err)
	}

	var amapResp amapRegeoResponse
	if err := json.Unmarshal(body, &amapResp); err != nil {
		return nil, fmt.Errorf("parse amap response: %w", err)
	}

	// 高德 status=1 表示成功
	if amapResp.Status != "1" {
		return nil, bizerror.New(CodeAmapError, amapResp.Info)
	}

	return &ReverseGeocodeResponse{
		Province: amapResp.Regeocode.AddressComponent.Province,
		City:     amapResp.Regeocode.AddressComponent.City,
		District: amapResp.Regeocode.AddressComponent.District,
	}, nil
}
