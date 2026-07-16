package store

import (
	"context"
	"encoding/json"
	"fmt"
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
			Name:     st.Name,
			Address:  buildAddress(st.RegionName, st.Address),
			Distance: distance,
			Unit:     unit,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Distance < items[j].Distance
	})

	return items, nil
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

// buildAddress 将 region_name（如 ["广东省","广州市","番禺区"]）与详细地址拼接
// 返回形如 "广东省广州市番禺区xxxxxx" 的完整地址
func buildAddress(regionName json.RawMessage, address string) string {
	if len(regionName) == 0 {
		return address
	}
	var names []string
	if err := json.Unmarshal(regionName, &names); err != nil {
		return address
	}
	return strings.Join(names, "") + address
}
