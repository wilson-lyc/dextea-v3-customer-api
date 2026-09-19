package menu

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/productrpc"
	productv1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/product/v1"
)

type Service struct {
	client *productrpc.Client
}

func NewService(client *productrpc.Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetStoreMenu(ctx context.Context, storeID int64) (*StoreMenuResponse, error) {
	resp, err := s.client.StoreMenu(ctx, &productv1.GetStoreMenuRequest{
		StoreId: uint64(storeID),
	})
	if err != nil {
		if bizErr, ok := bizerror.As(err); ok && bizErr.Code == productrpc.ErrNotFound.Code {
			return nil, bizerror.New(
				&bizerror.BizError{Code: 40400, Message: "资源不存在", Kind: bizerror.KindBusiness},
				"门店未配置菜单",
			)
		}
		return nil, err
	}
	if resp == nil || resp.Menu == nil || resp.Tree == nil {
		return nil, productrpc.ErrNotFound
	}

	out := &StoreMenuResponse{
		Menu:   MenuInfo{ID: int64(resp.Menu.Id), Name: resp.Menu.Name, Description: resp.Menu.Description},
		Groups: []GroupInfo{},
	}
	for _, group := range resp.Tree.Groups {
		if group == nil {
			continue
		}
		item := GroupInfo{ID: int64(group.GroupId), Name: group.Name, Sort: int(group.Sort), Products: []ProductInfo{}}
		for i, product := range group.Products {
			if product == nil {
				continue
			}
			item.Products = append(item.Products, ProductInfo{
				ID:     int64(product.ProductId),
				Name:   product.Name,
				Brief:  product.Brief,
				Price:  product.Price,
				Sort:   i,
				Status: int(product.Status),
			})
		}
		out.Groups = append(out.Groups, item)
	}
	return out, nil
}
