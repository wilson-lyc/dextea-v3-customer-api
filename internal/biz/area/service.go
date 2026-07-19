package area

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

const (
	amapRegeoURL    = "https://restapi.amap.com/v3/geocode/regeo"
	amapHTTPTimeout = 5 * time.Second
)

// 城市列表缓存键与过期时间。城市列表更新频率低，缓存 10 分钟足矣。
const citiesCacheKey = "dextea:store:cities"
const citiesCacheTTL = 10 * time.Minute

type Service struct {
	repo       *Repository
	rdb        *redis.Client
	httpClient *http.Client
	apiKey     string
}

func NewService(cfg *config.Config, repo *Repository, rdb *redis.Client) *Service {
	return &Service{
		repo: repo,
		rdb:  rdb,
		httpClient: &http.Client{
			Timeout: amapHTTPTimeout,
		},
		apiKey: cfg.AmapAPIKey,
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

// GetCities 获取城市列表
// 优先从 Redis 缓存读取；缓存未命中或 Redis 不可用时回源 MySQL，
// 命中后写回缓存以避免后续请求反复访问数据库。
func (s *Service) GetCities(ctx context.Context) ([]CityLetterGroup, error) {
	// 1. 尝试从 Redis 读取缓存
	if s.rdb != nil {
		if cached, err := s.rdb.Get(ctx, citiesCacheKey).Bytes(); err == nil && len(cached) > 0 {
			var result []CityLetterGroup
			if jsonErr := json.Unmarshal(cached, &result); jsonErr == nil {
				return result, nil
			}
		}
	}

	// 2. 缓存未命中（或不可用），回源数据库
	cities, err := s.repo.GetDistinctCities(ctx)
	if err != nil {
		return nil, err
	}

	if len(cities) == 0 {
		return []CityLetterGroup{}, nil
	}

	// 按拼音首字母分组
	groupMap := make(map[string][]string)
	for _, city := range cities {
		letter := cityFirstLetter(city)
		groupMap[letter] = append(groupMap[letter], city)
	}

	// 收集所有字母并排序
	letters := make([]string, 0, len(groupMap))
	for l := range groupMap {
		letters = append(letters, l)
	}
	sort.Strings(letters)

	// 构建结果
	result := make([]CityLetterGroup, 0, len(letters))
	for _, l := range letters {
		result = append(result, CityLetterGroup{
			Letter: l,
			Cities: groupMap[l],
		})
	}

	// 3. 写回缓存，供后续请求直接命中（Redis 不可用时静默跳过）。
	if s.rdb != nil {
		if data, jsonErr := json.Marshal(result); jsonErr == nil {
			_ = s.rdb.Set(ctx, citiesCacheKey, data, citiesCacheTTL).Err()
		}
	}

	return result, nil
}

// cityFirstLetter 获取城市名称的拼音首字母（小写），失败返回 "#"。
func cityFirstLetter(city string) string {
	runeCity := []rune(city)
	if len(runeCity) == 0 {
		return "#"
	}

	// 如果首字符是英文字母则直接返回小写
	first := runeCity[0]
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
		return strings.ToLower(string(first))
	}

	// 尝试拼音转换
	a := pinyin.NewArgs()
	a.Style = pinyin.FirstLetter
	result := pinyin.Pinyin(string(first), a)
	if len(result) > 0 && len(result[0]) > 0 {
		return strings.ToLower(result[0][0])
	}

	return "#"
}
