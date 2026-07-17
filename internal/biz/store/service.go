package store

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

const geoKey = "dextea:store:location"

type Service struct {
	repo *Repository
	rdb  *redis.Client
}

func NewService(repo *Repository, rdb *redis.Client) *Service {
	return &Service{
		repo: repo,
		rdb:  rdb,
	}
}

// Nearby 获取附近门店
func (s *Service) Nearby(ctx context.Context, req NearbyRequest) ([]NearbyStoreItem, error) {
	if s.rdb == nil {
		return nil, bizerror.New(CodeRedisDisabled)
	}
	if s.repo.db == nil {
		return nil, bizerror.New(CodeDBDisabled)
	}

	// 1. 通过 Redis GEO 获取附近门店 ID 及距离
	geoResults, err := s.rdb.GeoRadius(ctx, geoKey, req.Longitude, req.Latitude, &redis.GeoRadiusQuery{
		Radius:    req.Distance,
		Unit:      "km",
		WithDist:  true,
		WithCoord: false,
		Count:     req.Count,
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, err
	}

	if len(geoResults) == 0 {
		return []NearbyStoreItem{}, nil
	}

	// 2. 提取门店 ID 列表并保持距离映射
	idDistMap := make(map[int64]float64, len(geoResults))
	ids := make([]int64, 0, len(geoResults))
	for _, gr := range geoResults {
		id, err := parseStoreID(gr.Name)
		if err != nil {
			continue
		}
		ids = append(ids, id)
		idDistMap[id] = gr.Dist
	}

	if len(ids) == 0 {
		return []NearbyStoreItem{}, nil
	}

	// 3. 批量查询门店数据
	stores, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 4. 组装结果：按距离从近到远排序
	items := make([]NearbyStoreItem, 0, len(stores))
	for _, st := range stores {
		distKm := idDistMap[st.ID]
		distance, unit := formatDistance(distKm)
		items = append(items, NearbyStoreItem{
			ID:       st.ID,
			Name:     st.Name,
			Address:  buildAddress(st.Province, st.City, st.District, st.Address),
			Distance: distance,
			Unit:     unit,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Distance < items[j].Distance
	})

	return items, nil
}

// Search 按条件搜索门店。
// province/city/district 按省市区文本筛选，keyword 模糊匹配（名称或地址）；
// 依据传入的经纬度计算各门店距离，结果按由近到远排序。返回结构与 Nearby 一致。
func (s *Service) Search(ctx context.Context, req SearchRequest) ([]NearbyStoreItem, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(CodeDBDisabled)
	}

	stores, err := s.repo.Search(ctx, req.Province, req.City, req.District, req.Keyword)
	if err != nil {
		return nil, err
	}
	if len(stores) == 0 {
		return []NearbyStoreItem{}, nil
	}

	items := make([]NearbyStoreItem, 0, len(stores))
	for _, st := range stores {
		distKm := haversine(req.Longitude, req.Latitude, st.Longitude, st.Latitude)
		distance, unit := formatDistance(distKm)
		items = append(items, NearbyStoreItem{
			ID:       st.ID,
			Name:     st.Name,
			Address:  buildAddress(st.Province, st.City, st.District, st.Address),
			Distance: distance,
			Unit:     unit,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Distance < items[j].Distance
	})

	return items, nil
}

// haversine 计算两点间的大圆距离（单位：km）。
func haversine(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadiusKm = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// parseStoreID 将 Redis GEO member（门店 ID 字符串）解析为 int64
func parseStoreID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}

// formatDistance 将 km 距离格式化为合适的单位
// 距离 < 1km 返回 m，否则返回 km
func formatDistance(km float64) (float64, string) {
	if km < 1 {
		return km * 1000, "m"
	}
	return km, "km"
}

// buildAddress 将 省/市/区 文本与详细地址拼接
// 返回形如 "广东省广州市番禺区xxxxxx" 的完整地址
func buildAddress(province, city, district, address string) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{province, city, district} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "") + address
}
