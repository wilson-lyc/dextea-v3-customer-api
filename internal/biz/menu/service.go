package menu

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/productrpc"
	productv1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/product/v1"
)

type Service struct {
	repo   *Repository
	client *productrpc.Client
}

func NewService(repo *Repository, client *productrpc.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) GetStoreMenu(ctx context.Context, storeID int64) (*StoreMenuResponse, error) {
	menuID, err := s.repo.FindStoreMenuID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if menuID == 0 {
		return nil, bizerror.New(&bizerror.BizError{Code: 40400, Message: "资源不存在", Kind: bizerror.KindBusiness}, "门店未配置菜单")
	}

	menu, err := s.client.Menu(ctx, &productv1.GetMenuRequest{Id: uint64(menuID)})
	if err != nil {
		return nil, err
	}
	tree, err := s.client.MenuTree(ctx, &productv1.GetMenuTreeRequest{MenuId: uint64(menuID), Mode: productv1.MenuTreeMode_MENU_TREE_MODE_BUSINESS, StoreId: func() *uint64 { v := uint64(storeID); return &v }()})
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, productrpc.ErrNotFound
	}
	if menu == nil {
		return nil, productrpc.ErrNotFound
	}
	resp := &StoreMenuResponse{Menu: MenuInfo{ID: int64(menu.Id), Name: menu.Name, Description: menu.Description}, Groups: []GroupInfo{}}
	for _, group := range tree.Groups {
		if group == nil {
			continue
		}
		g := GroupInfo{ID: int64(group.GroupId), Name: group.Name, Sort: int(group.Sort), Products: []ProductInfo{}}
		for i, p := range group.Products {
			if p != nil {
				status := int(p.Status)
				g.Products = append(g.Products, ProductInfo{ID: int64(p.ProductId), Name: p.Name, Brief: p.Brief, Price: p.Price, Sort: i, Status: status})
			}
		}
		resp.Groups = append(resp.Groups, g)
	}
	return resp, nil
}
