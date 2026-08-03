package store

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/asmarques/geodist"
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

// 获取附近门店
func (s *Service) Nearby(ctx context.Context, req NearbyRequest) ([]StoreDetailItem, error) {
	if s.rdb == nil {
		return nil, bizerror.New(bizerror.ErrRedisDisabled)
	}
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	// 通过 Redis GEO 获取附近门店 ID 及距离
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
		return []StoreDetailItem{}, nil
	}

	// 提取门店 ID 列表并保持距离映射
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
		return []StoreDetailItem{}, nil
	}

	// 批量查询门店数据
	stores, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 组装结果：按距离从近到远排序
	items := make([]StoreDetailItem, 0, len(stores))
	for _, st := range stores {
		distKm := idDistMap[st.ID]
		distance, unit := formatDistance(distKm)
		dist := distance
		u := unit
		items = append(items, StoreDetailItem{
			ID:            st.ID,
			Name:          st.Name,
			Status:        st.Status,
			Province:      st.Province,
			City:          st.City,
			District:      st.District,
			Address:       buildAddress(st.Province, st.City, st.District, st.Address),
			BusinessHours: st.BusinessHours,
			Phone:         st.Phone,
			Longitude:     st.Longitude,
			Latitude:      st.Latitude,
			Distance:      &dist,
			Unit:          &u,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return *items[i].Distance < *items[j].Distance
	})

	return items, nil
}

// 搜索门店
func (s *Service) Search(ctx context.Context, req SearchRequest) ([]StoreDetailItem, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	stores, err := s.repo.Search(ctx, req.City, req.Keyword)
	if err != nil {
		return nil, err
	}
	if len(stores) == 0 {
		return []StoreDetailItem{}, nil
	}

	items := make([]StoreDetailItem, 0, len(stores))
	for _, st := range stores {
		distKm := haversine(req.Longitude, req.Latitude, st.Longitude, st.Latitude)
		distance, unit := formatDistance(distKm)
		dist := distance
		u := unit
		items = append(items, StoreDetailItem{
			ID:            st.ID,
			Name:          st.Name,
			Status:        st.Status,
			Province:      st.Province,
			City:          st.City,
			District:      st.District,
			Address:       buildAddress(st.Province, st.City, st.District, st.Address),
			BusinessHours: st.BusinessHours,
			Phone:         st.Phone,
			Longitude:     st.Longitude,
			Latitude:      st.Latitude,
			Distance:      &dist,
			Unit:          &u,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return *items[i].Distance < *items[j].Distance
	})

	return items, nil
}

// 获取门店详情
func (s *Service) GetDetail(ctx context.Context, req GetDetailRequest) (*StoreDetailItem, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	store, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, nil
	}

	distKm := haversine(req.Longitude, req.Latitude, store.Longitude, store.Latitude)
	distance, unit := formatDistance(distKm)
	dist := distance
	u := unit

	return &StoreDetailItem{
		ID:            store.ID,
		Name:          store.Name,
		Status:        store.Status,
		Province:      store.Province,
		City:          store.City,
		District:      store.District,
		Address:       buildAddress(store.Province, store.City, store.District, store.Address),
		BusinessHours: store.BusinessHours,
		Phone:         store.Phone,
		Longitude:     store.Longitude,
		Latitude:      store.Latitude,
		Distance:      &dist,
		Unit:          &u,
	}, nil
}

// 计算两点距离（单位 km），使用 geodist 包实现。
func haversine(lng1, lat1, lng2, lat2 float64) float64 {
	return geodist.HaversineDistance(
		geodist.Point{Lat: lat1, Long: lng1},
		geodist.Point{Lat: lat2, Long: lng2},
	)
}

// parseStoreID 将 Redis GEO member（门店 ID 字符串）解析为 int64
func parseStoreID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}

// 距离格式化
func formatDistance(km float64) (float64, string) {
	if km < 1 {
		return km * 1000, "m"
	}
	return km, "km"
}

// 地址格式化
func buildAddress(province, city, district, address string) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{province, city, district} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "") + address
}
