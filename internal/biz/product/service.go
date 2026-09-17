package product

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/infra/productrpc"
	productv1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/product/v1"
)

type Service struct {
	client *productrpc.Client
}

func NewService(client *productrpc.Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetDetail(ctx context.Context, req GetProductDetailRequest) (*ProductDetailResponse, error) {
	detail, err := s.client.Detail(ctx, &productv1.GetProductDetailRequest{ProductId: uint64(req.ProductID), StoreId: uint64(req.StoreID)})
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.Product == nil {
		return nil, productrpc.ErrNotFound
	}
	p := detail.Product
	result := &ProductDetailResponse{ID: int64(p.Id), Name: p.Name, Brief: p.Brief, Description: p.Description, Status: int(detail.StoreStatus), Price: p.Price, Gallery: []ProductImageItem{}, Customizations: []CustomizationItemResponse{}}
	if detail.Images != nil {
		if detail.Images.Cover != nil {
			result.Cover = &ProductImageItem{ID: int64(detail.Images.Cover.Id), URL: detail.Images.Cover.Url, Type: 1}
		}
		for i, image := range detail.Images.Gallery {
			result.Gallery = append(result.Gallery, ProductImageItem{ID: int64(image.Id), URL: image.Url, Sort: i, Type: 2})
		}
	}
	for _, item := range detail.CustomizationItems {
		if item == nil || item.Item == nil {
			continue
		}
		ci := CustomizationItemResponse{ID: int64(item.Item.Id), Name: item.Item.Name, Sort: int(item.Item.Sort), Status: int(item.Item.Status), Options: []CustomizationOptionItem{}}
		for _, option := range item.Options {
			if option != nil && option.Option != nil {
				ci.Options = append(ci.Options, CustomizationOptionItem{ID: int64(option.Option.Id), Name: option.Option.Name, Price: option.Option.Price, Sort: int(option.Option.Sort), Status: int(option.StoreStatus)})
			}
		}
		result.Customizations = append(result.Customizations, ci)
	}
	return result, nil
}

func (s *Service) GetStoreStatus(ctx context.Context, req GetProductStoreStatusRequest) ([]ProductStoreStatusItem, error) {
	ids := make([]uint64, 0, len(req.ProductIDs))
	for _, id := range req.ProductIDs {
		ids = append(ids, uint64(id))
	}
	response, err := s.client.StoreStatuses(ctx, &productv1.GetProductStoreStatusesRequest{StoreId: uint64(req.StoreID), ProductIds: ids})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return []ProductStoreStatusItem{}, nil
	}
	items := make([]ProductStoreStatusItem, 0, len(response.Products))
	for _, p := range response.Products {
		items = append(items, ProductStoreStatusItem{ProductID: int64(p.ProductId), Name: p.Name, Status: int(p.StoreStatus)})
	}
	return items, nil
}
