package store

import (
	"context"
	"strings"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/storerpc"
	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"
)

type Service struct {
	client *storerpc.Client
}

func NewService(client *storerpc.Client) *Service {
	return &Service{client: client}
}

func (s *Service) Nearby(ctx context.Context, req NearbyRequest) ([]StoreDetailItem, error) {
	resp, err := s.client.Nearby(ctx, &storev1.GetBusinessNearbyStoresRequest{
		Longitude:  req.Longitude,
		Latitude:   req.Latitude,
		DistanceKm: req.Distance,
		Count:      int32(req.Count),
	})
	if err != nil {
		return nil, err
	}
	items := make([]StoreDetailItem, 0, len(resp.GetStores()))
	for _, item := range resp.GetStores() {
		if item == nil || item.GetStore() == nil {
			continue
		}
		items = append(items, toDetailItem(item.GetStore(), item.GetDistanceKm()))
	}
	return items, nil
}

func (s *Service) Search(ctx context.Context, req SearchRequest) ([]StoreDetailItem, error) {
	resp, err := s.client.Search(ctx, &storev1.SearchBusinessStoresRequest{
		City:      req.City,
		Keyword:   req.Keyword,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		return nil, err
	}
	items := make([]StoreDetailItem, 0, len(resp.GetStores()))
	for _, item := range resp.GetStores() {
		if item == nil || item.GetStore() == nil {
			continue
		}
		items = append(items, toDetailItem(item.GetStore(), item.GetDistanceKm()))
	}
	return items, nil
}

func (s *Service) GetDetail(ctx context.Context, req GetDetailRequest) (*StoreDetailItem, error) {
	item, err := s.client.Detail(ctx, &storev1.GetBusinessStoreRequest{
		Id:        uint64(req.ID),
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		if isStoreNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if item == nil || item.GetStore() == nil {
		return nil, nil
	}
	out := toDetailItem(item.GetStore(), item.GetDistanceKm())
	return &out, nil
}

func toDetailItem(st *storev1.BusinessStore, distanceKm float64) StoreDetailItem {
	distance, unit := formatDistance(distanceKm)
	return StoreDetailItem{
		ID:            int64(st.GetId()),
		Name:          st.GetName(),
		Status:        int(st.GetStatus()),
		Province:      st.GetProvince(),
		City:          st.GetCity(),
		District:      st.GetDistrict(),
		Address:       buildAddress(st.GetProvince(), st.GetCity(), st.GetDistrict(), st.GetAddress()),
		BusinessHours: st.GetBusinessHours(),
		Phone:         st.GetPhone(),
		Longitude:     st.GetLongitude(),
		Latitude:      st.GetLatitude(),
		Distance:      &distance,
		Unit:          &unit,
	}
}

func isStoreNotFound(err error) bool {
	bizErr, ok := bizerror.As(err)
	return ok && bizErr.Code == storerpc.ErrNotFound.Code
}

func formatDistance(km float64) (float64, string) {
	if km < 1 {
		return km * 1000, "m"
	}
	return km, "km"
}

func buildAddress(province, city, district, address string) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{province, city, district} {
		if strings.TrimSpace(p) != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "") + address
}
